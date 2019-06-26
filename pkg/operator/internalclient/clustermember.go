package internalclient

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	internalv1client "github.com/openshift/cluster-etcd-operator/pkg/generated/clientset/versioned/typed/operator/v1"
	internalv1informers "github.com/openshift/cluster-etcd-operator/pkg/generated/informers/externalversions"
	internalv1 "github.com/openshift/pkg/apis/etcd.openshift.io/v1"
)

type ClusterMemberClient struct {
	Informers internalv1informers.SharedInformerFactory
	Client    internalv1client.ClusterMembersGetter
}

func (c *ClusterMemberClient) Informer() cache.SharedIndexInformer {
	return c.Informers.Etcd().V1().ClusterMembers().Informer()
}

func (c *ClusterMemberClient) GetStaticPodOperatorState() (*internalv1.ClusterMemberSpec, *internalv1.ClusterMemberStatus, string, error) {
	ClusterLister
	instance, err := c.Informers.Etcd().V1().ClusterMembers().Lister().Get("cluster")
	if err != nil {
		return nil, nil, "", err
	}

	return &instance.Spec.ClusterMemberSpec, &instance.Status.ClusterMemberStatus, instance.ResourceVersion, nil
}

func (c *ClusterMemberClient) GetStaticPodOperatorStateWithQuorum() (*internalv1.ClusterMemberSpec, *internalv1.ClusterMemberStatus, string, error) {
	instance, err := c.Client.Etcds().Get("cluster", metav1.GetOptions{})
	if err != nil {
		return nil, nil, "", err
	}

	return &instance.Spec.ClusterMemberSpec, &instance.Status.ClusterMemberStatus, instance.ResourceVersion, nil
}

func (c *ClusterMemberClient) UpdateClusterMemberSpec(resourceVersion string, spec *internalv1.ClusterMemberSpec) (*internalv1.ClusterMemberSpec, string, error) {
	original, err := c.Informers.Operator().V1().Etcds().Lister().Get("cluster")
	if err != nil {
		return nil, "", err
	}
	copy := original.DeepCopy()
	copy.ResourceVersion = resourceVersion
	copy.Spec.ClusterMemberSpec = *spec

	ret, err := c.Client.Etcds().Update(copy)
	if err != nil {
		return nil, "", err
	}

	return &ret.Spec.ClusterMemberSpec, ret.ResourceVersion, nil
}

func (c *ClusterMemberClient) UpdateClusterMemberStatus(resourceVersion string, status *internalv1.ClusterMemberStatus) (*internalv1.ClusterMemberStatus, error) {
	original, err := c.Informers.Operator().V1().Etcds().Lister().Get("cluster")
	if err != nil {
		return nil, err
	}
	copy := original.DeepCopy()
	copy.ResourceVersion = resourceVersion
	copy.Status.ClusterMemberStatus = *status

	ret, err := c.Client.Etcds().UpdateStatus(copy)
	if err != nil {
		return nil, err
	}

	return &ret.Status.ClusterMemberStatus, nil
}

func (c *ClusterMemberClient) GetOperatorState() (*internalv1.OperatorSpec, *internalv1.OperatorStatus, string, error) {
	instance, err := c.Informers.Operator().V1().Etcds().Lister().Get("cluster")
	if err != nil {
		return nil, nil, "", err
	}

	return &instance.Spec.OperatorSpec, &instance.Status.ClusterMemberStatus.OperatorStatus, instance.ResourceVersion, nil
}

func (c *ClusterMemberClient) UpdateOperatorSpec(resourceVersion string, spec *internalv1.OperatorSpec) (*internalv1.OperatorSpec, string, error) {
	original, err := c.Informers.Operator().V1().Etcds().Lister().Get("cluster")
	if err != nil {
		return nil, "", err
	}
	copy := original.DeepCopy()
	copy.ResourceVersion = resourceVersion
	copy.Spec.OperatorSpec = *spec

	ret, err := c.Client.Etcds().Update(copy)
	if err != nil {
		return nil, "", err
	}

	return &ret.Spec.OperatorSpec, ret.ResourceVersion, nil
}
func (c *ClusterMemberClient) UpdateOperatorStatus(resourceVersion string, status *internalv1.OperatorStatus) (*internalv1.OperatorStatus, error) {
	original, err := c.Informers.Operator().V1().Etcds().Lister().Get("cluster")
	if err != nil {
		return nil, err
	}
	copy := original.DeepCopy()
	copy.ResourceVersion = resourceVersion
	copy.Status.ClusterMemberStatus.OperatorStatus = *status

	ret, err := c.Client.Etcds().UpdateStatus(copy)
	if err != nil {
		return nil, err
	}

	return &ret.Status.ClusterMemberStatus.OperatorStatus, nil
}
