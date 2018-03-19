package k8s

import (
	"testing"

	"fmt"
	"strings"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/database"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
	k8sTesting "k8s.io/client-go/testing"
)

type expectedActions struct {
	namespace string
	verb      string
	resource  string
	name      string
}

func TestMetricsExporter_CreateMetricsExporter(t *testing.T) {
	f := fake.NewSimpleClientset()
	c := &MetricsExporter{
		clientset: f,
	}

	err := c.CreateMetricsExporter(database.Scope("test-shadow"), "test", database.CredentialID("test-shadow-test-monitoring"))
	assert.Nil(t, err)

	actions := f.Actions()
	assert.NotEmpty(t, actions)

}

func TestMetricsExporter_CreateMetricsExporterActions(t *testing.T) {

	e := []expectedActions{
		// Master secret initial save
		{namespace: "test-shadow", verb: "get", resource: "configmaps", name: "test-metrics-exporter"},
		{namespace: "test-shadow", verb: "create", resource: "configmaps"},
		{namespace: "test-shadow", verb: "get", resource: "services", name: "test-metrics-exporter"},
		{namespace: "test-shadow", verb: "create", resource: "services"},
		{namespace: "test-shadow", verb: "get", resource: "deployments", name: "test-metrics-exporter"},
		{namespace: "test-shadow", verb: "create", resource: "deployments"},
	}

	f := fake.NewSimpleClientset()
	c := &MetricsExporter{
		clientset: f,
	}

	err := c.CreateMetricsExporter(database.Scope("test-shadow"), "test", database.CredentialID("test-shadow-test-monitoring"))
	assert.Nil(t, err)

	actions := f.Actions()
	assert.NotEmpty(t, actions)
	assertActions(t, e, actions)

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
