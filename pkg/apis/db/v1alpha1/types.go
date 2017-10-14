package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DB is a specification for a DB resource
type DB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DBSpec   `json:"spec"`
	Status            DBStatus `json:"status"`
}

// DBSpec is the spec for a DB resource
type DBSpec struct {
	Type string `json:"type"`
}

// DBStatus is the status for a DB resource
type DBStatus struct {
	Ready string `json:"ready"`
}
