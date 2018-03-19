package worker_test

import (
	"testing"

	crds "github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/worker"

	"fmt"
	"strings"

	fake2 "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned/fake"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/database"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/k8s"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/mocks"
	"github.com/golang/mock/gomock"
	_ "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	k8sTesting "k8s.io/client-go/testing"
)

type expectedAction struct {
	namespace string
	verb      string
	resource  string
	name      string
}

func TestOnCreate_HappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	crd := crds.PostgresDB{}
	crd.ObjectMeta.Name = "crdname"
	crd.ObjectMeta.Namespace = "test-namespace"
	crd.ObjectMeta.UID = "2098284b-1daf-11e8-b83f-028cde27f28a"
	crd.Spec.Size = "db.m4.large"
	crd.Spec.Storage = 5

	// fake client set to assert actions
	f := fake.NewSimpleClientset()
	crdF := fake2.NewSimpleClientset()

	wrkr, retDBAvailable := getWorker(ctrl, crd, database.StatusAvailable, f, crdF)

	// Given
	expectedK8sActions := []expectedAction{

		// Master secret initial save
		{namespace: "kube-system", verb: "get", resource: "secrets", name: getCRDNameForCredential(crd.Namespace, crd.Name, retDBAvailable.Credentials[0].ID)},
		{namespace: "kube-system", verb: "create", resource: "secrets"},

		// Master secret update host info
		{namespace: "kube-system", verb: "get", resource: "secrets", name: getCRDNameForCredential(crd.Namespace, crd.Name, retDBAvailable.Credentials[0].ID)},
		{namespace: "kube-system", verb: "update", resource: "secrets"},

		// Application User secret save
		{namespace: "test-namespace", verb: "get", resource: "secrets", name: getCRDNameForCredential(crd.Namespace, crd.Name, retDBAvailable.Credentials[1].ID)},
		{namespace: "test-namespace", verb: "create", resource: "secrets"},

		// Application Admin secret save
		{namespace: "test-namespace", verb: "get", resource: "secrets", name: getCRDNameForCredential(crd.Namespace, crd.Name, retDBAvailable.Credentials[2].ID)},
		{namespace: "test-namespace", verb: "create", resource: "secrets"},

		// Readonly secret save
		{namespace: "test-namespace", verb: "get", resource: "secrets", name: getCRDNameForCredential(crd.Namespace, crd.Name, retDBAvailable.Credentials[3].ID)},
		{namespace: "test-namespace", verb: "create", resource: "secrets"},

		//System secret save
		{namespace: "test-namespace-shadow", verb: "get", resource: "secrets", name: getCRDNameForCredential(crd.Namespace, crd.Name, retDBAvailable.Credentials[4].ID)},
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

	wrkr.PostgresDBValidator.(*mocks.MockPostgresDBValidator).EXPECT().Validate(gomock.Any()).Return(nil).Times(1)
	wrkr.DBCreateGetter.(*mocks.MockDBCreateGetter).EXPECT().GetDB(gomock.Any()).Return(nil, nil).Times(1)
	wrkr.DBCreateGetter.(*mocks.MockDBCreateGetter).EXPECT().CreateDB(gomock.Any(), gomock.Any()).Return(retDBAvailable, nil).Times(1)
	wrkr.DBCreateGetter.(*mocks.MockDBCreateGetter).EXPECT().GetDB(gomock.Any()).Return(retDBAvailable, nil).Times(1)

	// When
	wrkr.OnCreate(&crd)

	// Then
	assertActions(t, expectedK8sActions, f.Actions())
}

func TestOnCreate_WrongCRD(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	crd := crds.PostgresDB{}
	crd.Namespace = "test"
	crd.Name = "test"
	crd.ObjectMeta.UID = "2098284b-1daf-11e8-b83f-028cde27f28a"

	crdF := fake2.NewSimpleClientset()
	crdF.PostgresdbV1alpha1().PostgresDBs(crd.Namespace).Create(&crd)

	dbWrkr, _ := getWorker(ctrl, crd, database.StatusAvailable, fake.NewSimpleClientset(), crdF)
	gomock.InOrder(
		dbWrkr.PostgresDBValidator.(*mocks.MockPostgresDBValidator).EXPECT().Validate(gomock.Any()).Return(fmt.Errorf("Something exploded")).Times(1),
	)
	dbWrkr.Logger.(*mocks.MockLogger).EXPECT().Error("invalid postgresdb object: Something exploded").Times(1)
	dbWrkr.OnCreate(&crd)

}

func isMatchingNamespace(e expectedAction, a k8sTesting.Action) bool {
	return e.namespace == a.GetNamespace()
}

func isMatchingVerbResource(e expectedAction, a k8sTesting.Action) bool {
	return a.Matches(e.verb, e.resource)
}

func isMatchingName(e expectedAction, a k8sTesting.Action) bool {
	return e.name == "" || containsField(a, "Name", e.name)
}

func isActionExpected(expecteds []expectedAction, a k8sTesting.Action) bool {
	for _, e := range expecteds {
		if isMatchingNamespace(e, a) &&
			isMatchingVerbResource(e, a) &&
			isMatchingName(e, a) {
			return true
		}
	}
	return false
}

func assertActions(t *testing.T, expecteds []expectedAction, actual []k8sTesting.Action) {
	if len(expecteds) != len(actual) {
		t.Fatalf("expected %d action(s): got(%d)[%s]", len(expecteds), len(actual), actual)
	}
	for _, action := range actual {
		//for i, action := range actual {
		//fmt.Printf("index %d, ACTION: %v %v %v\n", i, action.GetResource(), action.GetVerb(), action.GetNamespace())
		if !isActionExpected(expecteds, action) {
			t.Errorf("cannot find the action %v", action)
		}

	}
}

func containsField(object interface{}, fieldName, fieldValue string) bool {
	objectString := fmt.Sprintf("%#v", object)
	field := fmt.Sprintf("%s:%#v", fieldName, fieldValue)

	return strings.Contains(strings.ToLower(objectString), strings.ToLower(field))
}

func getCRDNameForCredential(namespace string, name string, id database.CredentialID) string {
	return fmt.Sprintf("%s-%s-%s", namespace, name, string(id))
}

func getWorker(ctrl *gomock.Controller, crd crds.PostgresDB, status database.Status, f *fake.Clientset, crdF *fake2.Clientset) (*worker.DBWorker, *database.Database) {

	config := worker.NewConfig(100, "shadow")
	crdF.PostgresdbV1alpha1().PostgresDBs(crd.Namespace).Create(&crd)

	c := k8s.NewStoreCreds(f)
	r := mocks.NewMockDBCreateGetter(ctrl)
	m := k8s.NewMetricsExporter(f)
	v := mocks.NewMockPostgresDBValidator(ctrl)
	l := mocks.NewMockLogger(ctrl)
	tfm := worker.NewOptimus()
	s := k8s.NewCRDClient(crdF)

	// retVals
	id := fmt.Sprintf("%s-%s-%s", crd.Namespace, crd.Name, crd.UID)

	creds := database.Credentials{
		database.CredTypeAdmin:       &database.Credential{ID: database.CredentialID("master")},
		database.CredTypeAppAdmin:    &database.Credential{ID: database.CredentialID("appadmin")},
		database.CredTypeAppUser:     &database.Credential{ID: database.CredentialID("appuser")},
		database.CredTypeAppReadOnly: &database.Credential{ID: database.CredentialID("appreadonly")},
		database.CredTypeMonitoring:  &database.Credential{ID: database.CredentialID("monitoring")},
	}
	retDBAvailable := &database.Database{
		Status:      status,
		ID:          database.DatabaseID(id),
		Name:        crd.Name,
		Credentials: creds,
	}

	wrkr := worker.NewDBWorker(r, c, m, config, v, l, tfm, s)
	return wrkr, retDBAvailable
}

func alwaysHappyCalls(wrkr *worker.DBWorker, retDB *database.Database) {
	wrkr.PostgresDBValidator.(*mocks.MockPostgresDBValidator).EXPECT().Validate(gomock.Any()).Return(nil).Times(1)
	wrkr.DBCreateGetter.(*mocks.MockDBCreateGetter).EXPECT().GetDB(gomock.Any()).Return(nil, nil).Times(1)
	wrkr.DBCreateGetter.(*mocks.MockDBCreateGetter).EXPECT().CreateDB(gomock.Any(), gomock.Any()).Return(retDB, nil).Times(1)
	wrkr.DBCreateGetter.(*mocks.MockDBCreateGetter).EXPECT().GetDB(gomock.Any()).Return(retDB, nil).Times(1)
}
