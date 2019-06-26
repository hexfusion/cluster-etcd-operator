package internalclient

import (
	internalv1client "github.com/openshift/cluster-etcd-operator/pkg/generated/clientset/versioned/typed/operator/v1"
	internalv1informers "github.com/openshift/cluster-etcd-operator/pkg/generated/informers/externalversions"
)

type ClusterClient struct {
	Informers internalv1informers.SharedInformerFactory
	Client    internalv1client.ClustersGetter
}

// TODO fill this in not going to worry about cluster yet.
