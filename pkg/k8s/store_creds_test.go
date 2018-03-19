package k8s

import (
	"testing"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/database"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetCreds_Found(t *testing.T) {

	fakeClient := fake.NewSimpleClientset()
	k := &StoreCreds{
		client: fakeClient,
	}

	fakeClient.CoreV1().Secrets("test").Create(getSecret())
	id := database.CredentialID("test")
	cred, err := k.GetCred(database.Scope("test"), id)
	assert.Nil(t, err)
	assert.NotNil(t, cred)

}
func TestGetCreds_NotFound(t *testing.T) {

	id := database.CredentialID("test")
	fakeClient := fake.NewSimpleClientset()
	k := &StoreCreds{client: fakeClient}

	cred, err := k.GetCred(database.Scope("test"), id)

	assert.Nil(t, err)
	assert.Nil(t, cred)

}

// TODO implement this test
//func TestGetCreds_Error() {
//
//}

func TestUpdateCreds_Success(t *testing.T) {

	id := database.CredentialID("test")
	cred := &database.Credential{ID: id, Scope: "test"}
	fakeClient := fake.NewSimpleClientset()

	k := &StoreCreds{client: fakeClient}

	fakeClient.CoreV1().Secrets("test").Create(getSecret())

	err := k.UpdateCred(cred)

	assert.Nil(t, err)

}

func TestUpdateCreds_Error(t *testing.T) {

	id := database.CredentialID("test")
	cred := &database.Credential{ID: id, Scope: "test"}
	fakeClient := fake.NewSimpleClientset()

	k := &StoreCreds{client: fakeClient}

	err := k.UpdateCred(cred)

	assert.NotNil(t, err)

}

func TestCreateCreds_Success(t *testing.T) {

	id := database.CredentialID("test")
	cred := &database.Credential{ID: id, Scope: "test"}
	fakeClient := fake.NewSimpleClientset()

	k := &StoreCreds{client: fakeClient}

	fakeClient.CoreV1().Secrets("test").Create(getSecret())

	err := k.UpdateCred(cred)

	assert.Nil(t, err)

}

func TestCreateCreds_Error(t *testing.T) {

	id := database.CredentialID("test")
	cred := &database.Credential{ID: id, Scope: "test"}
	fakeClient := fake.NewSimpleClientset()

	k := &StoreCreds{client: fakeClient}

	err := k.UpdateCred(cred)

	assert.NotNil(t, err)

}

func getSecret() *v1.Secret {
	var secret = make(map[string][]byte)
	secret["DB_HOST"] = []byte("banana")
	secret["DB_USER"] = []byte("test")
	return &v1.Secret{
		ObjectMeta: v12.ObjectMeta{
			Name:            "test",
			Namespace:       "test",
			OwnerReferences: []v12.OwnerReference{},
		},
		Data: secret,
	}
}
