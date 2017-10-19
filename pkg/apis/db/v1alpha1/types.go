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
	Name string `json:"name"`
	Size string `json:"size,omitempty"`
	//	Port          int64  `json:"port,omitempty"`
	//	Subnet        string `json:"subnet"`
	//	SecurityGroup string `json:"securityGroup"`
	Multizone bool `json:"multizone,omitempty"`
	//	Username      string `json:"username,omitempty"`
	//	Password      string `json:"password,omitempty"`
	//	EncryptionKey string `json:"encryptionKey,omitempty"`
}

// DBStatus is the status for a DB resource
type DBStatus struct {
	Ready string `json:"ready"`
	ARN   string `json:"arn"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DBList is a list of DB resources
type DBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DB `json:"items"`
}
