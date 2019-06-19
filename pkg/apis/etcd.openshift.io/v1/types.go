package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Member struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec contains the request itself
	Spec MemberSpec `json:"spec"`

	// Status contains information about the request
	// + optional
	Status MemberStatus `json:"status,omitempty"`
}

type MemberSpec struct {
	// Name of the member
	Name string `json:"name"`
	// PeerURLs of the member
	PeerURLs string `json:"peerURLs"`
}

type MemberStatus struct {
	// Conditions applied to the request, such as approval or denial.
	// +optional
	Conditions []MemberCondition `json:"conditions,omitempty"`

	// Config is returned if request is approved.
	Config []byte `json:"config,omitempty"`
}

type MemberCondition struct {
	// member state success/failure.
	Type string `json:"type"`
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
