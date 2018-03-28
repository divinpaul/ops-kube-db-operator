package worker

//go:generate mockgen -source=$GOFILE -destination=../mocks/mock_validator.go -package=mocks

import (
	"fmt"

	"strconv"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/rds"
)

type PostgresDBValidator interface {
	Validate(crd *v1alpha1.PostgresDB) error
}

type postgresDBvalidator struct{}

func NewPostgresDBValidator() *postgresDBvalidator {
	return &postgresDBvalidator{}
}

func (v *postgresDBvalidator) Validate(crd *v1alpha1.PostgresDB) error {
	if crd.Spec.Storage == "" {
		return fmt.Errorf("storage cannot be empty")
	}

	_, err := strconv.ParseInt(crd.Spec.Storage, 10, 64)
	if err != nil {
		return err
	}

	if crd.Spec.Size == "" {
		return fmt.Errorf("size cannot be empty")
	}

	if _, err = rds.GetSizeForInstanceClass(crd.Spec.Size); err != nil {
		return fmt.Errorf("unsupported database size")
	}

	return nil
}
