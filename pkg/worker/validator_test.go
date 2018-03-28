package worker

import (
	"testing"

	"github.com/golang/mock/gomock"

	crds "github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestValidate_StorageEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	crd := crds.PostgresDB{}
	crd.ObjectMeta.Name = "crdname"
	crd.ObjectMeta.Namespace = "test-namespace"
	crd.Spec.Size = "db.m4.large"
	crd.Spec.Storage = ""

	i := NewPostgresDBValidator()
	err := i.Validate(&crd)

	assert.NotNil(t, err)

}

func TestValidate_StorageInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	crd := crds.PostgresDB{}
	crd.ObjectMeta.Name = "crdname"
	crd.ObjectMeta.Namespace = "test-namespace"
	crd.Spec.Size = "db.m4.large"
	crd.Spec.Storage = "banana"

	i := NewPostgresDBValidator()
	err := i.Validate(&crd)

	assert.NotNil(t, err)

}

func TestValidate_SizeEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	crd := crds.PostgresDB{}
	crd.ObjectMeta.Name = "crdname"
	crd.ObjectMeta.Namespace = "test-namespace"
	crd.Spec.Size = ""
	crd.Spec.Storage = "10"

	i := NewPostgresDBValidator()
	err := i.Validate(&crd)

	assert.NotNil(t, err)

}

func TestValidate_SizeInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	crd := crds.PostgresDB{}
	crd.ObjectMeta.Name = "crdname"
	crd.ObjectMeta.Namespace = "test-namespace"
	crd.Spec.Size = "nonexistent"
	crd.Spec.Storage = "10"

	i := NewPostgresDBValidator()
	err := i.Validate(&crd)

	assert.NotNil(t, err)

}
