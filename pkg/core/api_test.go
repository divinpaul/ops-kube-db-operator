package core

import (
	"testing"

	"fmt"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/database"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreateDatabaseIfNotExist_GetDBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i, _, req, cred := getCreateDBIfNotExistsScenario(ctrl)

	i.(*mocks.MockDBCreateGetter).EXPECT().GetDB(req.ID).Return(nil, fmt.Errorf("error")).Times(1)

	db, err := CreateDatabaseIfNotExist(i, req, cred)
	assert.NotNil(t, err)
	assert.Nil(t, db)
}

func TestCreateDatabaseIfNotExist_GetDBReturnExisting(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i, retDB, req, cred := getCreateDBIfNotExistsScenario(ctrl)

	i.(*mocks.MockDBCreateGetter).EXPECT().GetDB(req.ID).Return(retDB, nil).Times(1)

	db, err := CreateDatabaseIfNotExist(i, req, cred)
	assert.NotNil(t, db)
	assert.Nil(t, err)
}

func TestCreateDatabaseIfNotExist_CreateDBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i, _, req, cred := getCreateDBIfNotExistsScenario(ctrl)

	i.(*mocks.MockDBCreateGetter).EXPECT().GetDB(req.ID).Return(nil, nil).Times(1)
	i.(*mocks.MockDBCreateGetter).EXPECT().CreateDB(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).Times(1)

	db, err := CreateDatabaseIfNotExist(i, req, cred)
	assert.NotNil(t, err)
	assert.Nil(t, db)
}

func TestCreateDatabaseIfNotExistHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i, retDB, req, cred := getCreateDBIfNotExistsScenario(ctrl)
	i.(*mocks.MockDBCreateGetter).EXPECT().GetDB(req.ID).Return(nil, nil).Times(1)
	i.(*mocks.MockDBCreateGetter).EXPECT().CreateDB(gomock.Any(), gomock.Any()).Return(retDB, nil).Times(1)

	db, err := CreateDatabaseIfNotExist(i, req, cred)
	assert.Nil(t, err)
	assert.Equal(t, db, retDB)
}

// WaitForDBToAvailable

func TestWaitForDBToBeAvailable_StraightAway(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i := mocks.NewMockDBGetter(ctrl)
	id := database.DatabaseID("test1")
	retDB := getReturnDB(id, database.StatusAvailable)

	i.EXPECT().GetDB(id).Return(retDB, nil).Times(1)

	db, err := WaitForDBToBeAvailable(i, id, 1)
	assert.Nil(t, err)
	assert.Equal(t, db, retDB)
}

func TestWaitForDBToBeAvailable_AfterOneMili(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i := mocks.NewMockDBGetter(ctrl)
	id := database.DatabaseID("test1")
	retDBUnavailable := getReturnDB(id, database.StatusUnavailable)
	retDBAvailable := getReturnDB(id, database.StatusAvailable)

	gomock.InOrder(
		i.EXPECT().GetDB(id).Return(retDBUnavailable, nil).Times(9),
		i.EXPECT().GetDB(id).Return(retDBAvailable, nil).Times(1),
	)
	db, err := WaitForDBToBeAvailable(i, id, 5)

	assert.Nil(t, err)
	assert.Equal(t, db, retDBAvailable)
}

func TestWaitForDBToBeAvailable_NotBeforeTimeout(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i := mocks.NewMockDBGetter(ctrl)
	id := database.DatabaseID("test1")
	retDBUnavailable := getReturnDB(id, database.StatusUnavailable)

	i.EXPECT().GetDB(id).Return(retDBUnavailable, nil).Times(10)

	db, err := WaitForDBToBeAvailable(i, id, 5)
	assert.NotNil(t, err)
	assert.Nil(t, db)
}

func TestWaitForDBToBeAvailable_GetDBError(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i := mocks.NewMockDBGetter(ctrl)
	id := database.DatabaseID("test1")

	i.EXPECT().GetDB(id).Return(nil, fmt.Errorf("error")).Times(1)

	db, err := WaitForDBToBeAvailable(i, id, 1)
	assert.NotNil(t, err)
	assert.Nil(t, db)
}

// STORE DB Credentials Tests

func TestStoreDBCredentials_GetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	credAdmin := &database.Credential{ID: "test1", Scope: "kube-system"}

	creds := &database.Credentials{
		database.CredTypeAdmin: credAdmin,
	}

	i := mocks.NewMockCredentialsStorer(ctrl)

	i.EXPECT().GetCred(credAdmin.Scope, credAdmin.ID).Return(nil, fmt.Errorf("error")).Times(1)
	i.EXPECT().CreateCred(credAdmin).Times(0)

	err := StoreDBCredentials(i, creds)

	assert.NotNil(t, err)
}

func TestStoreDBCredentials_CreateSuccessfully(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	credAdmin := &database.Credential{ID: "test1", Scope: "kube-system"}
	credAppAdmin := &database.Credential{ID: "test2", Scope: "test"}

	creds := &database.Credentials{
		database.CredTypeAdmin:    credAdmin,
		database.CredTypeAppAdmin: credAppAdmin,
	}

	i := mocks.NewMockCredentialsStorer(ctrl)

	i.EXPECT().GetCred(credAdmin.Scope, credAdmin.ID).Return(nil, nil).Times(1)
	i.EXPECT().CreateCred(credAdmin).Return(nil).Times(1)
	i.EXPECT().GetCred(credAppAdmin.Scope, credAppAdmin.ID).Return(nil, nil).Times(1)
	i.EXPECT().CreateCred(credAppAdmin).Return(nil).Times(1)

	err := StoreDBCredentials(i, creds)

	assert.Nil(t, err)
}
func TestStoreDBCredentials_CreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	credAdmin := &database.Credential{ID: "test1", Scope: "kube-system"}

	creds := &database.Credentials{
		database.CredTypeAdmin: credAdmin,
	}

	i := mocks.NewMockCredentialsStorer(ctrl)

	i.EXPECT().GetCred(credAdmin.Scope, credAdmin.ID).Return(nil, nil).Times(1)
	i.EXPECT().CreateCred(credAdmin).Return(fmt.Errorf("error")).Times(1)

	err := StoreDBCredentials(i, creds)

	assert.NotNil(t, err)
}

func TestStoreDBCredentials_UpdateSuccessfully(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	credAdmin := &database.Credential{ID: "test1", Scope: "kube-system"}
	credAppAdmin := &database.Credential{ID: "test2", Scope: "test"}
	credAppReadOnly := &database.Credential{ID: "test3", Scope: "test"}

	creds := &database.Credentials{
		database.CredTypeAdmin:       credAdmin,
		database.CredTypeAppAdmin:    credAppAdmin,
		database.CredTypeAppReadOnly: credAppReadOnly,
	}

	i := mocks.NewMockCredentialsStorer(ctrl)

	i.EXPECT().GetCred(credAdmin.Scope, credAdmin.ID).Return(credAdmin, nil).Times(1)
	i.EXPECT().UpdateCred(credAdmin).Return(nil).Times(1)
	i.EXPECT().GetCred(credAppAdmin.Scope, credAppAdmin.ID).Return(credAppAdmin, nil).Times(1)
	i.EXPECT().UpdateCred(credAppAdmin).Return(nil).Times(1)
	i.EXPECT().GetCred(credAppReadOnly.Scope, credAppReadOnly.ID).Return(credAppReadOnly, nil).Times(1)
	i.EXPECT().UpdateCred(credAppReadOnly).Return(nil).Times(1)

	err := StoreDBCredentials(i, creds)

	assert.Nil(t, err)
}

func TestStoreDBCredentials_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	credAdmin := &database.Credential{ID: "test1", Scope: "kube-system"}

	creds := &database.Credentials{
		database.CredTypeAdmin: credAdmin,
	}

	i := mocks.NewMockCredentialsStorer(ctrl)

	i.EXPECT().GetCred(credAdmin.Scope, credAdmin.ID).Return(credAdmin, nil).Times(1)
	i.EXPECT().UpdateCred(credAdmin).Return(fmt.Errorf("error")).Times(1)

	err := StoreDBCredentials(i, creds)

	assert.NotNil(t, err)
}

func getCreateDBIfNotExistsScenario(ctrl *gomock.Controller) (DBCreateGetter, *database.Database, *database.Request, *database.Credential) {
	i := mocks.NewMockDBCreateGetter(ctrl)
	db := &database.Database{
		ID: database.DatabaseID("test1"),
	}
	r := &database.Request{
		Size:    database.SizeXSmall,
		Storage: 5,
		Name:    "test",
		Owner:   "test",
		HA:      false,
		ID:      "banana",
	}
	c := &database.Credential{
		CredType: database.CredTypeAdmin,
		Username: "banana",
		Password: "crypticbanana",
	}
	return i, db, r, c
}

func getReturnDB(id database.DatabaseID, status database.Status) *database.Database {
	return &database.Database{
		ID:     id,
		Status: status,
	}
}
