package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PostgresDB is a specification for a DB resource
type PostgresDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PostgresDBSpec   `json:"spec"`
	Status            PostgresDBStatus `json:"status"`
}

// PostgresDBSpec is the spec for a DB resource
type PostgresDBSpec struct {
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

// PostgresDBStatus is the status for a DB resource
type PostgresDBStatus struct {
	Ready string `json:"ready"`
	ARN   string `json:"arn"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PostgresDBList is a list of DB resources
type PostgresDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PostgresDB `json:"items"`
}
