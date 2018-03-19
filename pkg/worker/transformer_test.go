package worker

import (
	"testing"

	"fmt"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestCRDToRequest_DBIDSize(t *testing.T) {
	crd := &v1alpha1.PostgresDB{}
	crd.Namespace = "common-ledger-migrations"
	crd.Name = "bankfeed-migrator"
	crd.UID = "2098284b-1daf-11e8-b83f-028cde27f28a"
	crd.Spec.Size = "db.t2.small"
	crd.Spec.Storage = 5

	optimus := NewOptimus()
	req := optimus.CRDToRequest(crd)

	assert.NotNil(t, req)
	if len(req.ID) > 64 {
		t.Fail()
	}
}

func TestCRDToRequest_HappyPath(t *testing.T) {

	crd := &v1alpha1.PostgresDB{}
	crd.Namespace = "test-ns"
	crd.Name = "test"
	crd.UID = "2098284b-1daf-11e8-b83f-028cde27f28a"
	crd.Spec.Size = "db.t2.small"
	crd.Spec.Storage = 5
	crd.Spec.HA = true
	tags := map[string]string{
		"test2":      "test2",
		"owner":      crd.Namespace,
		"crd-name":   crd.Name,
		"created-by": "ops-kube-db-operator",
		"test1":      "test1",
	}

	crd.Spec.Tags = tags

	optimus := NewOptimus()
	req := optimus.CRDToRequest(crd)

	assert.NotNil(t, req)
	assert.Equal(t, req.ID, database.DatabaseID(fmt.Sprintf("%s-%s-%s", crd.Namespace, crd.Name, crd.GetUID())))
	assert.Equal(t, req.Metadata, tags)
}
