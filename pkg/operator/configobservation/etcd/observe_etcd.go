package etcd

import (
	"fmt"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"

	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/events"

	"github.com/openshift/cluster-etcd-operator/pkg/operator/clustermembercontroller"
	"github.com/openshift/cluster-etcd-operator/pkg/operator/configobservation"
)

const (
	etcdEndpointNamespace = "openshift-etcd"
	etcdHostEndpointName  = "host-etcd"
	etcdEndpointName      = "etcd"
)

// type etcdObserver struct {
// 	observerFunc  observeAPIServerConfigFunc
// 	configPaths   [][]string
// 	resourceNames []string
// 	resourceType  interface{}
// }

// TODO break out logic into functions to reduce dupe code.
// ObserveClusterMembers observes the current etcd cluster members.
func ObserveClusterMembers(genericListers configobserver.Listers, recorder events.Recorder, existingConfig map[string]interface{}) (map[string]interface{}, []error) {
	listers := genericListers.(configobservation.Listers)
	clusterMemberPath := []string{"cluster", "members"}
	observedConfig := map[string]interface{}{}

	var errs []error

	previouslyObservedMembers, found, err := unstructured.NestedSlice(existingConfig, clusterMemberPath...)
	if err != nil {
		errs = append(errs, err)
	}
	if found {
		if err := unstructured.SetNestedSlice(observedConfig, previouslyObservedMembers, clusterMemberPath...); err != nil {
			errs = append(errs, err)
		}
	}

	previousMemberCount := len(previouslyObservedMembers)

	var etcdURLs []interface{}
	etcdHostEndpoints, err := listers.OpenshiftEtcdEndpointsLister.Endpoints(etcdEndpointNamespace).Get(etcdHostEndpointName)
	if errors.IsNotFound(err) {
		recorder.Warningf("ObserveClusterMembers", "Required %s/%s endpoint not found", etcdEndpointNamespace, etcdHostEndpointName)
		return nil, append(errs, fmt.Errorf("endpoints/host-etcd.openshift-etcd: not found"))
	}
	if err != nil {
		recorder.Warningf("ObserveClusterMembers", "Error getting %s/%s endpoint: %v", etcdEndpointNamespace, etcdHostEndpointName, err)
		return nil, append(errs, err)
	}
	dnsSuffix := etcdHostEndpoints.Annotations["alpha.installer.openshift.io/dns-suffix"]
	if len(dnsSuffix) == 0 {
		dnsErr := fmt.Errorf("endpoints %s/%s: alpha.installer.openshift.io/dns-suffix annotation not found", etcdEndpointNamespace, etcdHostEndpointName)
		recorder.Warning("ObserveClusterMembersFailed", dnsErr.Error())
		return nil, append(errs, dnsErr)
	}
	currentMemberCount := 0
	// handle bootstrap etcd
	for _, subset := range etcdHostEndpoints.Subsets {
		for _, address := range subset.Addresses {
			if address.Hostname == "etcd-bootstrap" {
				etcdURL := map[string]interface{}{}
				name := address.Hostname
				if err := unstructured.SetNestedField(etcdURL, name, "name"); err != nil {
					return existingConfig, append(errs, err)
				}
				peerURLs := fmt.Sprintf("https://%s.%s:2380", name, dnsSuffix)
				if err := unstructured.SetNestedField(etcdURL, peerURLs, "peerURLs"); err != nil {
					return existingConfig, append(errs, err)
				}
				currentMemberCount++
				etcdURLs = append(etcdURLs, etcdURL)
			}
		}
	}

	// TODO handle flapping if the member was listed and is now not available then we keep the old value.
	// membership removal requires a seperate observastion method, perhaps metrics. A finalizer could handle
	// removal but we also need to be aware that an admin can jump in do what they please.
	etcdEndpoints, err := listers.OpenshiftEtcdEndpointsLister.Endpoints(etcdEndpointNamespace).Get(etcdEndpointName)
	if errors.IsNotFound(err) {
		recorder.Warningf("ObserveClusterMembers", "Required %s/%s endpoint not found", etcdEndpointNamespace, etcdEndpointName)
		return nil, append(errs, fmt.Errorf("endpoints/etcd.openshift-etcd: not found"))
	}
	if err != nil {
		recorder.Warningf("ObserveClusterMembers", "Error getting %s/%s endpoint: %v", etcdEndpointNamespace, etcdEndpointName, err)
		return nil, append(errs, err)
	}
	for _, subset := range etcdEndpoints.Subsets {
		for _, address := range subset.Addresses {
			etcdURL := map[string]interface{}{}
			name := address.TargetRef.Name
			if err := unstructured.SetNestedField(etcdURL, name, "name"); err != nil {
				return existingConfig, append(errs, err)
			}

			peerURLs := fmt.Sprintf("https://%s:2380", address.IP)
			if err := unstructured.SetNestedField(etcdURL, peerURLs, "peerURLs"); err != nil {
				return existingConfig, append(errs, err)
			}
			currentMemberCount++
			etcdURLs = append(etcdURLs, etcdURL)
		}
	}
	previousMembers, err := getMemberListFromConfig(previouslyObservedMembers)
	if err != nil {
		errs = append(errs, err)
	}

	var observerdCrashloop bool
	if currentMemberCount >= previousMemberCount {
		if err := unstructured.SetNestedField(observedConfig, etcdURLs, clusterMemberPath...); err != nil {
			klog.Warningf("errors writing observedConfig using old: %+v", errs)
			return existingConfig, append(errs, err)
		}
	} else {
		for _, previousMember := range previousMembers {
			if previousMember.Name != "etcd-bootstrap" {
				etcdPod, err := listers.OpenshiftEtcdPodsLister.Pods(etcdEndpointNamespace).Get(previousMember.Name)
				if errors.IsNotFound(err) {
					// verify the node exists
					//TODO this is very opnionated could this come from the endpoint?
					nodeName := strings.TrimPrefix(previousMember.Name, "etcd-member-")

					//TODO this should be a function
					node, err := listers.NodeLister.Get(nodeName)
					if errors.IsNotFound(err) {
						// if the node is no londer available we use the endpoint observatiopn
						klog.Warningf("error: Node %s not found: writing observed endpoints to config %+v", etcdURLs, err)
						if err := unstructured.SetNestedField(observedConfig, etcdURLs, clusterMemberPath...); err != nil {
							klog.Warningf("errors writing observedConfig using old: %+v", errs)
							return existingConfig, append(errs, err)
						}
					}
					if len(node.Status.Conditions) > 0 {
						// switch
						if node.Status.Conditions[0].Type != "NodeStatusUnknown" || node.Status.Conditions[0].Type != "NodeStatusDown" {
							klog.Warningf("Node Condition not expected %s:", node.Status.Conditions[0].Type)
							// we dont know why node is not ready but we cant assume we want to scale it
							break
						}
					}

					recorder.Warningf("ObserveClusterMembers", "Pod %s listed as member not found use last", previousMember.Name)
					return existingConfig, append(errs, err)
				}

				// since the pod exists lets figure out if the endpoint being down is a result of etcd crashlooping
				if etcdPod.Status.ContainerStatuses[0].State.Waiting != nil && etcdPod.Status.ContainerStatuses[0].State.Waiting.Reason == "CrashLoopBackOff" {
					observerdCrashloop = true
				}
			}
		}
		if observerdCrashloop {
			// if the previous status was crashloop then we are going to scale down.
			klog.Warningf("Pod is crashlooping using exising config")

			if err := unstructured.SetNestedField(observedConfig, etcdURLs, clusterMemberPath...); err != nil {
				klog.Warningf("errors writing observedConfig using old: %+v", errs)
				return existingConfig, append(errs, err)
			}
			return observedConfig, errs
		}
		// for now we don't allow the list to deciment because we are only handling bootstrap
		// in future this needs proper consideration.
		recorder.Warningf("ObserveClusterMembers", "Possible flapping current members observed (%d) is less than previous (%v)", currentMemberCount, previousMemberCount)
		return existingConfig, errs

	}

	if len(errs) > 0 {
		return observedConfig, errs
	}

	if !reflect.DeepEqual(previouslyObservedMembers, etcdURLs) {
		recorder.Eventf("ObserveClusterMembersUpdated", "Updated cluster members to %v", etcdURLs)
	}
	return observedConfig, errs
}

// ObservePendingClusterMembers observes pending etcd cluster members who are atempting to join the cluster.
// TODO it is possible that a member which is part of the cluster can show pending status if Pod goes down. We need to handle this flapping.
// If you are member you should not renturn to pending. During bootstrap this isn't a fatal flaw but its not optimal.
func ObservePendingClusterMembers(genericListers configobserver.Listers, recorder events.Recorder, existingConfig map[string]interface{}) (map[string]interface{}, []error) {
	listers := genericListers.(configobservation.Listers)
	observedConfig := map[string]interface{}{}
	clusterMemberPath := []string{"cluster", "pending"}
	var errs []error

	currentClusterMembers, found, err := unstructured.NestedSlice(existingConfig, clusterMemberPath...)
	if err != nil {
		errs = append(errs, err)
	}
	if found {
		if err := unstructured.SetNestedSlice(observedConfig, currentClusterMembers, clusterMemberPath...); err != nil {
			errs = append(errs, err)
		}
	}

	var etcdURLs []interface{}
	etcdEndpoints, err := listers.OpenshiftEtcdEndpointsLister.Endpoints(etcdEndpointNamespace).Get(etcdEndpointName)
	if errors.IsNotFound(err) {
		recorder.Warningf("ObservePendingClusterMembers", "Required %s/%s endpoint not found", etcdEndpointNamespace, etcdEndpointName)
		klog.Warningf("using observed: %+v", observedConfig)
		return observedConfig, append(errs, fmt.Errorf("endpoints/etcd.openshift-etcd: not found"))
	}
	if err != nil {
		recorder.Warningf("ObservePendingClusterMembers", "Error getting %s/%s endpoint: %v", etcdEndpointNamespace, etcdEndpointName, err)
		klog.Warningf("using observed: %+v", observedConfig)
		return observedConfig, errs
	}
	for _, subset := range etcdEndpoints.Subsets {
		for _, address := range subset.NotReadyAddresses {
			etcdURL := map[string]interface{}{}
			name := address.TargetRef.Name
			if err := unstructured.SetNestedField(etcdURL, name, "name"); err != nil {
				return existingConfig, append(errs, err)
			}

			peerURLs := fmt.Sprintf("https://%s:2380", address.IP)
			if err := unstructured.SetNestedField(etcdURL, peerURLs, "peerURLs"); err != nil {
				return existingConfig, append(errs, err)
			}
			etcdURLs = append(etcdURLs, etcdURL)
		}
	}

	if len(etcdURLs) > 0 {
		if err := unstructured.SetNestedField(observedConfig, etcdURLs, clusterMemberPath...); err != nil {

			klog.Warningf("using existing: %+v", existingConfig)
			return existingConfig, append(errs, err)
		}
	}

	if len(errs) > 0 {
		return observedConfig, errs
	}

	if !reflect.DeepEqual(currentClusterMembers, etcdURLs) {
		recorder.Eventf("ObservePendingClusterMembersUpdated", "Updated pending cluster members to %v", etcdURLs)
	}
	return existingConfig, errs
}

// ObserveStorageURLs observes the storage config URLs. If there is a problem observing the current storage config URLs,
// then the previously observed storage config URLs will be re-used.
func ObserveStorageURLs(genericListers configobserver.Listers, recorder events.Recorder, currentConfig map[string]interface{}) (observedConfig map[string]interface{}, errs []error) {
	listers := genericListers.(configobservation.Listers)
	observedConfig = map[string]interface{}{}
	storageConfigURLsPath := []string{"storageConfig", "urls"}

	currentEtcdURLs, found, err := unstructured.NestedStringSlice(currentConfig, storageConfigURLsPath...)
	if err != nil {
		errs = append(errs, err)
	}
	if found {
		if err := unstructured.SetNestedStringSlice(observedConfig, currentEtcdURLs, storageConfigURLsPath...); err != nil {
			errs = append(errs, err)
		}
	}

	var etcdURLs []string
	etcdEndpoints, err := listers.OpenshiftEtcdEndpointsLister.Endpoints(etcdEndpointNamespace).Get(etcdHostEndpointName)
	if errors.IsNotFound(err) {
		recorder.Warningf("ObserveStorageFailed", "Required %s/%s endpoint not found", etcdEndpointNamespace, etcdHostEndpointName)
		errs = append(errs, fmt.Errorf("endpoints/host-etcd.openshift-etcd: not found"))
		return
	}
	if err != nil {
		recorder.Warningf("ObserveStorageFailed", "Error getting %s/%s endpoint: %v", etcdEndpointNamespace, etcdHostEndpointName, err)
		errs = append(errs, err)
		return
	}
	dnsSuffix := etcdEndpoints.Annotations["alpha.installer.openshift.io/dns-suffix"]
	if len(dnsSuffix) == 0 {
		dnsErr := fmt.Errorf("endpoints %s/%s: alpha.installer.openshift.io/dns-suffix annotation not found", etcdEndpointNamespace, etcdHostEndpointName)
		recorder.Warning("ObserveStorageFailed", dnsErr.Error())
		errs = append(errs, dnsErr)
		return
	}
	for subsetIndex, subset := range etcdEndpoints.Subsets {
		for addressIndex, address := range subset.Addresses {
			if address.Hostname == "" {
				addressErr := fmt.Errorf("endpoints %s/%s: subsets[%v]addresses[%v].hostname not found", etcdHostEndpointName, etcdEndpointNamespace, subsetIndex, addressIndex)
				recorder.Warningf("ObserveStorageFailed", addressErr.Error())
				errs = append(errs, addressErr)
				continue
			}
			etcdURLs = append(etcdURLs, "https://"+address.Hostname+"."+dnsSuffix+":2379")
		}
	}

	if len(etcdURLs) == 0 {
		emptyURLErr := fmt.Errorf("endpoints %s/%s: no etcd endpoint addresses found", etcdEndpointNamespace, etcdHostEndpointName)
		recorder.Warning("ObserveStorageFailed", emptyURLErr.Error())
		errs = append(errs, emptyURLErr)
	}

	if len(errs) > 0 {
		return
	}

	if err := unstructured.SetNestedStringSlice(observedConfig, etcdURLs, storageConfigURLsPath...); err != nil {
		errs = append(errs, err)
		return
	}

	if !reflect.DeepEqual(currentEtcdURLs, etcdURLs) {
		recorder.Eventf("ObserveStorageUpdated", "Updated storage urls to %s", strings.Join(etcdURLs, ","))
	}

	return
}

//TOFO dupe logic
func getMemberListFromConfig(config []interface{}) ([]clustermembercontroller.Member, error) {
	var members []clustermembercontroller.Member
	for _, member := range config {
		memberMap, _ := member.(map[string]interface{})
		name, exists, err := unstructured.NestedString(memberMap, "name")
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("member name does not exist")
		}
		peerURLs, exists, err := unstructured.NestedString(memberMap, "peerURLs")
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("member peerURLs do not exist")
		}

		m := clustermembercontroller.Member{
			Name:     name,
			PeerURLS: []string{peerURLs},
		}
		members = append(members, m)
	}
	return members, nil
}

// // extractPreviouslyObservedConfig extracts the previously observed config from the existing config.
// func extractPreviouslyObservedConfig(existing map[string]interface{}, paths ...string) (map[string]interface{}, []error) {
// 	var errs []error
// 	previous := map[string]interface{}{}
// 	for _, fields := range paths {
// 		value, found, err := unstructured.NestedFieldCopy(existing, fields...)
// 		if !found {
// 			continue
// 		}
// 		if err != nil {
// 			errs = append(errs, err)
// 		}
// 		err = unstructured.SetNestedField(previous, value, fields...)
// 		if err != nil {
// 			errs = append(errs, err)
// 		}
// 	}
// 	return previous, errs
// }
