package worker_test

import (
	"errors"
	"testing"

	crds "github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	fakeCrd "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned/fake"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/rds"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/worker"

	"fmt"
	"strings"

	_ "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	k8sTesting "k8s.io/client-go/testing"
)

type mockDBInstanceCreator struct {
	rds.DBInstanceCreator
	input             *rds.CreateInstanceInput
	output            *rds.CreateInstanceOutput
	shouldErrorCreate bool
}

func (m *mockDBInstanceCreator) Create(input *rds.CreateInstanceInput) (*rds.CreateInstanceOutput, error) {
	if m.shouldErrorCreate {
		return nil, errors.New("test error")
	}
	m.input = input
	return m.output, nil
}

var (
	defaultMockDBInstanceCreator = &mockDBInstanceCreator{shouldErrorCreate: false}
	defaultRdsConfig             = &worker.RDSConfig{}
)

func TestCreateFunction(t *testing.T) {
	// Given
	defaultMockCrds := fakeCrd.NewSimpleClientset()
	expectedCrdActions := []expectedActions{
		{namespace: "test-namespace", verb: "update", resource: "postgresdbs"},
		{namespace: "test-namespace", verb: "update", resource: "postgresdbs"},
	}

	defaultMockClientSet := fake.NewSimpleClientset()
	expectedK8sActions := []expectedActions{
		// Master secret initial save
		{namespace: "kube-system", verb: "get", resource: "secrets", name: "crdname-master"},
		{namespace: "kube-system", verb: "get", resource: "secrets", name: "crdname-master"},
		{namespace: "kube-system", verb: "create", resource: "secrets"},
		// Master secret update with rds instances
		{namespace: "kube-system", verb: "get", resource: "secrets", name: "crdname-master"},
		{namespace: "kube-system", verb: "get", resource: "secrets", name: "crdname-master"},
		{namespace: "kube-system", verb: "update", resource: "secrets"},
		// Admin secret save
		{namespace: "test-namespace", verb: "get", resource: "secrets", name: "crdname-admin"},
		{namespace: "test-namespace", verb: "get", resource: "secrets", name: "crdname-admin"},
		{namespace: "test-namespace", verb: "create", resource: "secrets"},
		// Metrics Exporter secret save
		{namespace: "test-namespace-shadow", verb: "get", resource: "secrets", name: "crdname-metrics-exporter"},
		{namespace: "test-namespace-shadow", verb: "get", resource: "secrets", name: "crdname-metrics-exporter"},
		{namespace: "test-namespace-shadow", verb: "create", resource: "secrets"},
		// Metrics Exporter config map save
		{namespace: "test-namespace-shadow", verb: "get", resource: "configmaps", name: "crdname-metrics-exporter"},
		{namespace: "test-namespace-shadow", verb: "create", resource: "configmaps"},
		// Metrics Exporter service save
		{namespace: "test-namespace-shadow", verb: "get", resource: "services", name: "crdname-metrics-exporter"},
		{namespace: "test-namespace-shadow", verb: "create", resource: "services"},
		// Metrics Exporter deployment save
		{namespace: "test-namespace-shadow", verb: "get", resource: "deployments", name: "crdname-metrics-exporter"},
		{namespace: "test-namespace-shadow", verb: "create", resource: "deployments"},
	}

	defaultMockDBInstanceCreator.output = &rds.CreateInstanceOutput{ARN: "test-arn"}

	wrkr := worker.NewRDSWorker(defaultMockDBInstanceCreator, defaultMockClientSet, defaultMockCrds.PostgresdbV1alpha1(), defaultRdsConfig)
	crd := crds.PostgresDB{}
	crd.ObjectMeta.Name = "crdname"
	crd.ObjectMeta.Namespace = "test-namespace"

	// When
	wrkr.OnCreate(&crd)

	// Then
	assertActions(t, expectedCrdActions, defaultMockCrds.Actions())
	assertActions(t, expectedK8sActions, defaultMockClientSet.Actions())

	if crd.Status.Ready != "available" {
		t.Errorf("CRD Status Ready not updated properly: %#v", crd.Status)
	}

	if crd.Status.ARN != "test-arn" {
		t.Errorf("CRD Status ARN not updated properly: %#v", crd.Status)
	}
}

func TestCreateFunctionWithCreateDBError(t *testing.T) {
	// Given
	defaultMockCrds := fakeCrd.NewSimpleClientset()
	expectedCrdActions := []expectedActions{
		{namespace: "test-namespace", verb: "update", resource: "postgresdbs"},
		{namespace: "test-namespace", verb: "update", resource: "postgresdbs"},
	}

	defaultMockClientSet := fake.NewSimpleClientset()
	expectedK8sActions := []expectedActions{
		{namespace: "kube-system", verb: "get", resource: "secrets", name: "crdname-master"},
		{namespace: "kube-system", verb: "get", resource: "secrets", name: "crdname-master"},
		{namespace: "kube-system", verb: "create", resource: "secrets"},
	}

	wrkr := worker.NewRDSWorker(&mockDBInstanceCreator{shouldErrorCreate: true}, defaultMockClientSet, defaultMockCrds.PostgresdbV1alpha1(), defaultRdsConfig)
	crd := crds.PostgresDB{}
	crd.ObjectMeta.Name = "crdname"
	crd.ObjectMeta.Namespace = "test-namespace"

	// when
	wrkr.OnCreate(&crd)

	// then
	assertActions(t, expectedCrdActions, defaultMockCrds.Actions())
	assertActions(t, expectedK8sActions, defaultMockClientSet.Actions())

	if crd.Status.Ready != "Error creating database instance: test error" {
		t.Errorf("CRD Status Ready not updated properly: %#v", crd.Status)
	}
}

type expectedActions struct {
	namespace string
	verb      string
	resource  string
	name      string
}

func assertActions(t *testing.T, expected []expectedActions, actual []k8sTesting.Action) {
	if len(expected) != len(actual) {
		t.Fatalf("expected %d action(s): got(%d)[%s]", len(expected), len(actual), actual)
	}

	for i, action := range actual {
		expectedAction := expected[i]
		if expectedAction.namespace != action.GetNamespace() {
			t.Errorf("Expected namespace:%s, got:%s for action[%d]: %s", expectedAction.namespace, action.GetNamespace(), i, action)
		}

		if !action.Matches(expectedAction.verb, expectedAction.resource) {
			t.Errorf("Expected verb:%s resource:%s to match for action[%d]: %s", expectedAction.verb, expectedAction.resource, i, action)
		}

		if expectedAction.name != "" && !containsField(action, "Name", expectedAction.name) {
			t.Errorf("Expected action to have resource with Name:%s for action[%d]: %#v", expectedAction.name, i, action)
		}
	}
}

func containsField(object interface{}, fieldName, fieldValue string) bool {
	objectString := fmt.Sprintf("%#v", object)
	field := fmt.Sprintf("%s:%#v", fieldName, fieldValue)

	return strings.Contains(strings.ToLower(objectString), strings.ToLower(field))
}
