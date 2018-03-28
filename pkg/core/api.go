package core

import (
	"fmt"
	"time"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/database"
)

//go:generate mockgen -source=$GOFILE -destination=../mocks/mock_core.go -package=mocks

// DBCreateor creates the database only
type DBCreator interface {
	CreateDB(req *database.Request, adminCred *database.Credential) (*database.Database, error)
}

// DBGetter checks if a Database already exists for the request
type DBGetter interface {
	GetDB(database.DatabaseID) (*database.Database, error)
}

// Gets credential
type CredsGetter interface {
	GetCred(credScope database.Scope, id database.CredentialID) (*database.Credential, error)
}

type CredsUpdater interface {
	UpdateCred(credential *database.Credential) error
}

type CredsCreator interface {
	CreateCred(credential *database.Credential) error
}

type MetricsExporterCreator interface {
	CreateMetricsExporter(s database.Scope, name string, id database.CredentialID) error
}

type StatusUpdater interface {
	StatusUpdate(sReq *database.StatusRequest) error
}

type CredentialsStorer interface {
	CredsCreator
	CredsGetter
	CredsUpdater
}

type DBCreateGetter interface {
	DBCreator
	DBGetter
}

type CreateDatabase interface {
	CredentialsStorer
	DBCreateGetter
}

func CreateDatabaseIfNotExist(i DBCreateGetter, req *database.Request, cred *database.Credential) (*database.Database, error) {

	// check if database already exists
	db, err := i.GetDB(req.ID)
	if err != nil {
		return nil, err
	}

	if db != nil {
		return db, nil
	}

	// create the database with master credential
	db, err = i.CreateDB(req, cred)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func WaitForDBToBeAvailable(i DBGetter, id database.DatabaseID, checkIntervalMillis int) (*database.Database, error) {

	numberOfChecks := 10
	tick := time.Tick(time.Duration(checkIntervalMillis) * time.Millisecond)
	timeout := time.After(time.Duration(numberOfChecks*checkIntervalMillis) * time.Millisecond)

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("database unavailable in %d milliseconds", numberOfChecks*checkIntervalMillis)
		case <-tick:
			db, err := i.GetDB(id)
			if err != nil {
				return nil, err
			}
			ok := verifyDBStatus(db.Status)
			if ok {
				return db, nil
			}

		}
	}

}

func StoreDBCredentials(i CredentialsStorer, creds *database.Credentials) error {

	for _, cred := range *creds {

		cr, err := i.GetCred(cred.Scope, cred.ID)

		if err != nil {
			// if error exists return nil, err
			return err
		} else if cr == nil {
			// if error is nil and creds are also nil
			// that means creds don't exist Create.

			err = i.CreateCred(cred)
			if err != nil {
				return err
			}
		} else {
			// if error is nil but creds are not nil
			// that means creds exist so Update.
			err = i.UpdateCred(cred)
			if err != nil {
				return err
			}

		}

	}
	return nil
}

func CreateMetricsExporterForDB(i MetricsExporterCreator, s database.Scope, name string, id database.CredentialID) error {
	return i.CreateMetricsExporter(s, name, id)
}

func UpdateStatus(i StatusUpdater, sReq *database.StatusRequest) error {
	return i.StatusUpdate(sReq)
}

func verifyDBStatus(status database.Status) bool {
	if status != database.StatusAvailable {
		return false
	}
	return true
}
