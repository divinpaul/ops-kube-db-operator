package worker_test

import (
	"testing"

	crds "github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	fakeCrd "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned/fake"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/db"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/worker"

	_ "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

type mockManager struct {
	db.RDSManager
}

var (
	defaultMockCrds      = fakeCrd.NewSimpleClientset()
	defaultMockClientSet = fake.NewSimpleClientset()
	defaultMockMgr       = &mockManager{}
)

func TestCreateFunction(t *testing.T) {
	wrkr := worker.NewRDSWorker(defaultMockMgr, defaultMockClientSet, defaultMockCrds.PostgresdbV1alpha1())
	crd := crds.PostgresDB{}
	crd.ObjectMeta.Name = "crdname"
	crd.ObjectMeta.Namespace = "test-namespace"

	wrkr.OnCreate(crd)
}
