package worker_test

import (
	"errors"
	"testing"

	crds "github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	fakeCrd "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned/fake"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/worker"

	_ "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

type mockInstanceCommands struct {
	rds.InstanceCommands
	createdDB         *db.DB
	shouldErrorCreate bool
}

func (m *mockManager) Create(db *db.DB, def *db.DB) (*db.DB, error) {
	if m.shouldErrorCreate {
		return nil, errors.New("test error")
	}
	arn := "test-arn"
	m.createdDB = db
	db.ARN = &arn
	return db, nil
}

var (
	defaultMockCrds      = fakeCrd.NewSimpleClientset()
	defaultMockClientSet = fake.NewSimpleClientset()
	defaultMockMgr       = &mockManager{shouldErrorCreate: false}
	defaultRdsConfig     = &worker.RDSConfig{}
)

func TestCreateFunction(t *testing.T) {
	defaultMockCrds.ClearActions()
	defaultMockClientSet.ClearActions()

	wrkr := worker.NewRDSWorker(defaultMockMgr, defaultMockClientSet, defaultMockCrds.PostgresdbV1alpha1(), defaultRdsConfig)
	crd := crds.PostgresDB{}
	crd.ObjectMeta.Name = "crdname"
	crd.ObjectMeta.Namespace = "test-namespace"

	wrkr.OnCreate(&crd)

	crdActions := defaultMockCrds.Actions()
	k8sActions := defaultMockClientSet.Actions()

	if len(crdActions) != 2 {
		t.Errorf("There are not 2 crd actions: %s", crdActions)
	}

	if len(k8sActions) != 8 {
		t.Errorf("There are not 8 k8s actions: %s", k8sActions)
	}
}

func TestCreateFunctionWithCreateDBError(t *testing.T) {
	defaultMockCrds.ClearActions()
	defaultMockClientSet.ClearActions()

	wrkr := worker.NewRDSWorker(&mockManager{shouldErrorCreate: true}, defaultMockClientSet, defaultMockCrds.PostgresdbV1alpha1(), defaultRdsConfig)
	crd := crds.PostgresDB{}
	crd.ObjectMeta.Name = "crdname"
	crd.ObjectMeta.Namespace = "test-namespace"

	wrkr.OnCreate(&crd)

	crdActions := defaultMockCrds.Actions()
	k8sActions := defaultMockClientSet.Actions()

	if len(crdActions) != 1 {
		t.Errorf("There are not 1 crd actions: %s", crdActions)
	}

	if len(k8sActions) != 4 {
		t.Errorf("There are not 4 k8s actions: %s", k8sActions)
	}

	if crd.Status.Ready != "Error creating Database: test error" {
		t.Errorf("CRD Status not updated properly: %s", crd)
	}
}