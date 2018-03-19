package worker

//go:generate mockgen -source=$GOFILE -destination=../mocks/mock_transformer.go -package=mocks

import (
	"fmt"

	"unicode/utf8"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/database"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/rds"
)

type Transformer interface {
	CRDToRequest(crd *v1alpha1.PostgresDB) *database.Request
}

type Optimus struct{}

func NewOptimus() *Optimus {
	return &Optimus{}
}

type Classes map[string]database.Size

func (o *Optimus) CRDToRequest(crd *v1alpha1.PostgresDB) *database.Request {

	crdName := crd.Name
	crdNS := crd.Namespace
	name := truncateBytes(fmt.Sprintf("%s-%s-%s", crdNS, crdName, crd.GetUID()), 63)
	dbID := database.DatabaseID(name)

	size, _ := rds.GetSizeForInstanceClass(crd.Spec.Size)

	req := &database.Request{
		ID:      dbID,
		Owner:   crd.Namespace,
		Name:    crdName,
		Size:    *size,
		Storage: crd.Spec.Storage,
		Metadata: map[string]string{
			"owner":      crdNS,
			"crd-name":   crdName,
			"created-by": "ops-kube-db-operator",
		},
	}

	if crd.Spec.HA {
		req.HA = crd.Spec.HA
	}

	if len(crd.Spec.Tags) > 0 {
		for k, v := range crd.Spec.Tags {
			req.Metadata[k] = v
		}
	}

	return req
}

func truncateBytes(s string, n int) string {
	for len(s) > n {
		_, i := utf8.DecodeLastRuneInString(s)
		s = s[:len(s)-i]
	}
	return s
}
