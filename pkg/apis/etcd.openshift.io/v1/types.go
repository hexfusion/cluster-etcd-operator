package v1

import (
	"github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CustomResourceDefinition for ClusterMember
// apiVersion: apiextensions.k8s.io/v1beta1
// kind: CustomResourceDefinition
// metadata:
//   # name must match the spec fields below, and be in the form: <plural>.<group>
//   name: clustermembers.etcd.openshift.io
// spec:
//   # group name to use for REST API: /apis/<group>/<version>
//   group: etcd.openshift.io
//   # list of versions supported by this CustomResourceDefinition
//   versions:
//     - name: v1
//       # Each version can be enabled/disabled by Served flag.
//       served: true
//       # One and only one version must be marked as the storage version.
//       storage: true
//   # either Namespaced or Cluster
//   scope: Namespaced
//   names:
//     # plural name to be used in the URL: /apis/<group>/<version>/<plural>
//     plural: clustermembers
//     # singular name to be used as an alias on the CLI and for display
//     singular: clustermember
//     # kind is normally the CamelCased singular type. Your resource manifests use this.
//     kind: ClusterMember
//     # shortNames allow shorter string to match your resource on the CLI
//     shortNames: clm
//

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterMember struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec contains the request itself
	Spec ClusterMemberSpec `json:"spec"`

	// Status contains information about the request
	// + optional
	Status ClusterMemberStatus `json:"status,omitempty"`
}

type ClusterMemberSpec struct {
	// Name of the member
	Name string `json:"name"`
	// PeerURLs of the member
	PeerURLs []string `json:"peerURLs"`
}

type ClusterMemberStatus struct {
	// Conditions applied to the request, such as approval or denial.
	// +optional
	Conditions []ClusterMemberCondition `json:"conditions,omitempty"`

	// Config is returned if request is approved.
	Config []byte `json:"config,omitempty"`
}

type ClusterMemberCondition struct {
	// member state possible values are Added Exists Failure.
	Type ClusterMemberConditionType `json:"type"`
	// brief reason for the request state
	// +optional
	Reason string `json:"reason,omitempty"`
	// human readable message with details about the request state
	// +optional
	Message string `json:"message,omitempty"`
	// timestamp for the last update to this condition
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

// ClusterMemberConditionType valid conditions of a clustermember
type ClusterMemberConditionType string

const (
	ClusterMemberApproved ClusterMemberConditionType = "Approved"
	ClusterMemberDenied                              = "Denied"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterMemberList is a list of etcd members
type ClusterMemberList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ClusterMember `json:"items"`
}

// CustomResourceDefinition for Cluster
// apiVersion: apiextensions.k8s.io/v1beta1
// kind: CustomResourceDefinition
// metadata:
//   # name must match the spec fields below, and be in the form: <plural>.<group>
//   name: clusters.etcd.openshift.io
// spec:
//   # group name to use for REST API: /apis/<group>/<version>
//   group: etcd.openshift.io
//   # list of versions supported by this CustomResourceDefinition
//   versions:
//     - name: v1
//       # Each version can be enabled/disabled by Served flag.
//       served: true
//       # One and only one version must be marked as the storage version.
//       storage: true
//   # either Namespaced or Cluster
//   scope: Namespaced
//   names:
//     # plural name to be used in the URL: /apis/<group>/<version>/<plural>
//     plural: clusters
//     # singular name to be used as an alias on the CLI and for display
//     singular: cluster
//     # kind is normally the CamelCased singular type. Your resource manifests use this.
//     kind: Cluster
//     # shortNames allow shorter string to match your resource on the CLI
//     shortNames: cl
//

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cluster
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ClusterSpec   `json:"spec"`
	Status            ClusterStatus `json:"status"`
}

// ClusterSpec
type ClusterSpec struct {
	// Size is the expected size of the etcd cluster.
	// We should populate this with an observer.
	Size int `json:"size"`
	// Members is a list of expected members
	Members []string
}

// ClusterStatus
type ClusterStatus struct {
	// Condition keeps track of all cluster conditions, if they exist.
	Conditions []ClusterCondition `json:"conditions,omitempty"`

	// Size is the current size of the cluster
	Size int `json:"size"`

	// Members are the etcd members in the cluster
	Members MembersStatus `json:"members"`
}

// ClusterCondition represents one current condition of an etcd cluster.
// A condition might not show up if it is not happening.
// For example, if a cluster is not upgrading, the Upgrading condition would not show up.
// If a cluster is upgrading and encountered a problem that prevents the upgrade,
// the Upgrading condition's status will would be False and communicate the problem back.
type ClusterCondition struct {
	// Type of cluster condition.
	Type ClusterConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

type ClusterConditionType string

const (
	ClusterConditionAvailable  ClusterConditionType = "Available"
	ClusterConditionRecovering                      = "Recovering"
	ClusterConditionScaling                         = "Scaling"
	ClusterConditionUpgrading                       = "Upgrading"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterList is a list of etcd members
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Cluster `json:"items"`
}

// MembersStatus
type MembersStatus struct {
	// Ready are the etcd members that are ready to serve requests
	// The member names are the same as the etcd pod names
	Ready []string `json:"ready,omitempty"`
	// Unready are the etcd members not ready to serve requests
	Unready []string `json:"unready,omitempty"`
}
