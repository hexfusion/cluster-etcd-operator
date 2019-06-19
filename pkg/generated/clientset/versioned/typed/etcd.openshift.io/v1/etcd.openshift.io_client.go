// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/openshift/cluster-etcd-operator/pkg/apis/etcd.openshift.io/v1"
	"github.com/openshift/cluster-etcd-operator/pkg/generated/clientset/versioned/scheme"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type EtcdV1Interface interface {
	RESTClient() rest.Interface
	ClustersGetter
	ClusterMembersGetter
}

// EtcdV1Client is used to interact with features provided by the etcd.openshift.io group.
type EtcdV1Client struct {
	restClient rest.Interface
}

func (c *EtcdV1Client) Clusters() ClusterInterface {
	return newClusters(c)
}

func (c *EtcdV1Client) ClusterMembers() ClusterMemberInterface {
	return newClusterMembers(c)
}

// NewForConfig creates a new EtcdV1Client for the given config.
func NewForConfig(c *rest.Config) (*EtcdV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &EtcdV1Client{client}, nil
}

// NewForConfigOrDie creates a new EtcdV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *EtcdV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new EtcdV1Client for the given RESTClient.
func New(c rest.Interface) *EtcdV1Client {
	return &EtcdV1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *EtcdV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
