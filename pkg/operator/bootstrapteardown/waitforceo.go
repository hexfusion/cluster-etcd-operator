package bootstrapteardown

import (
	"context"

	operatorv1 "github.com/openshift/api/operator/v1"
	operatorversionedclient "github.com/openshift/client-go/operator/clientset/versioned"
	operatorv1helpers "github.com/openshift/library-go/pkg/operator/v1helpers"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	clientwatch "k8s.io/client-go/tools/watch"
	"k8s.io/klog"
)

func WaitForEtcdBootstrap(ctx context.Context, config *rest.Config) error {
	operatorConfigClient, err := operatorversionedclient.NewForConfig(config)
	if err != nil {
		return err
	}

	_, err = clientwatch.UntilWithSync(
		ctx,
		cache.NewListWatchFromClient(operatorConfigClient.OperatorV1().RESTClient(), "etcds", "", fields.OneTermEqualSelector("metadata.name", "cluster")),
		&operatorv1.Etcd{},
		nil,
		func(event watch.Event) (bool, error) {
			switch event.Type {
			case watch.Added, watch.Modified:
				etcd, ok := event.Object.(*operatorv1.Etcd)
				if !ok {
					klog.Warningf("Expected an Etcd object but got a %q object instead", event.Object.GetObjectKind().GroupVersionKind())
					return false, nil
				}
				if etcd.Spec.ManagementState == operatorv1.Unmanaged {
					klog.Info("Cluster etcd operator is in Unmanaged mode")
					return true, nil
				}
				if operatorv1helpers.IsOperatorConditionTrue(etcd.Status.Conditions, operatorv1.OperatorStatusTypeAvailable) &&
					operatorv1helpers.IsOperatorConditionFalse(etcd.Status.Conditions, operatorv1.OperatorStatusTypeProgressing) &&
					operatorv1helpers.IsOperatorConditionTrue(etcd.Status.Conditions, operatorv1.OperatorStatusTypeDegraded) {
					klog.Info("Cluster etcd operator bootstrapped successfully")
					return true, nil
				}
				klog.Infof("Still waiting for the cluster-etcd-operator to bootstrap")
				return false, nil
			}
			klog.Infof("Still waiting for the cluster-etcd-operator to bootstrap...")
			return false, nil
		},
	)

	if err == nil {
		klog.Infof("cluster-etcd-operator bootstrap etcd")
		return nil
	}
	return err
}
