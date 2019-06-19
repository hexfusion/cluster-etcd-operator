// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	time "time"

	etcdopenshiftiov1 "github.com/openshift/cluster-etcd-operator/pkg/apis/etcd.openshift.io/v1"
	versioned "github.com/openshift/cluster-etcd-operator/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/openshift/cluster-etcd-operator/pkg/generated/informers/externalversions/internalinterfaces"
	v1 "github.com/openshift/cluster-etcd-operator/pkg/generated/listers/etcd.openshift.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// MemberInformer provides access to a shared informer and lister for
// Members.
type MemberInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.MemberLister
}

type memberInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewMemberInformer constructs a new informer for Member type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewMemberInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredMemberInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredMemberInformer constructs a new informer for Member type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredMemberInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.EtcdV1().Members().List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.EtcdV1().Members().Watch(options)
			},
		},
		&etcdopenshiftiov1.Member{},
		resyncPeriod,
		indexers,
	)
}

func (f *memberInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredMemberInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *memberInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&etcdopenshiftiov1.Member{}, f.defaultInformer)
}

func (f *memberInformer) Lister() v1.MemberLister {
	return v1.NewMemberLister(f.Informer().GetIndexer())
}
