/*

Copyright 2017 MYOB Technology Pty Ltd

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
documentation files (the "Software"), to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

*/

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
	Size    string            `json:"size,omitempty"`
	Storage string            `json:"storage,omitempty"`
	Iops    int64             `json:"iops,omitempty"`
	HA      bool              `json:"ha,omitempty"`
	Tags    map[string]string `json:"tags,omitempty"`
}

// PostgresDBStatus is the status for a DB resource
type PostgresDBStatus struct {
	Ready string `json:"ready"`
	ARN   string `json:"arn"`
	ID    string `json:"id"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=postgresdbs
// PostgresDBList is a list of DB resources
type PostgresDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PostgresDB `json:"items"`
}
