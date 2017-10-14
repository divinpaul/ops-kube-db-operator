package v1alpha1

import (
	"github.com/gugahoi/rds-operator/pkg/apis/db"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// SchemeGroupVersion is group version used to register the objects
	SchemeGroupVersion = schema.GroupVersion{Group: db.GroupName, Version: "v1alpha1"}
	// SchemeBuilder is something
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
	// AddToScheme is another something
	AddToScheme = localSchemeBuilder.AddToScheme
)

// Resource tajes an unqualified resource and resturns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func init() {
	localSchemeBuilder.Register(addKnownTypes)
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion, &DB{})
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
