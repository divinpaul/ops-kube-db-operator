package worker

import (
	"fmt"

	crds "github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/core"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/database"
)

type DBWorker struct {
	*DBWorkerConfig
	PostgresDBValidator
	Transformer
	Logger
	core.DBCreateGetter
	core.StatusUpdater
	core.CredentialsStorer
	core.MetricsExporterCreator
}

type DBWorkerConfig struct {
	checkIntervalInMillis int
	nsSuffix              string
}

// NewRDSWorker returns new DBWorker instance for handling change events on postgresDB crd
func NewDBWorker(
	r core.DBCreateGetter,
	c core.CredentialsStorer,
	m core.MetricsExporterCreator,
	cfg *DBWorkerConfig,
	p PostgresDBValidator,
	l Logger,
	t Transformer,
	u core.StatusUpdater,
) *DBWorker {

	return &DBWorker{
		PostgresDBValidator:    p,
		DBWorkerConfig:         cfg,
		DBCreateGetter:         r,
		CredentialsStorer:      c,
		MetricsExporterCreator: m,
		Logger:                 l,
		Transformer:            t,
		StatusUpdater:          u,
	}
}

func NewConfig(c int, s string) *DBWorkerConfig {
	return &DBWorkerConfig{
		checkIntervalInMillis: c,
		nsSuffix:              s,
	}
}

func (w *DBWorker) OnCreate(obj interface{}) {

	crd := obj.(*crds.PostgresDB)
	s := database.Scope(crd.Namespace)

	// Validate the CRD
	if err := w.Validate(crd); err != nil {
		w.Error(fmt.Sprintf("invalid postgresdb object: %v", err))
		updateCRDStatus(w.StatusUpdater, w.Logger, crd.Name, s, database.StatusErrored, nil)
		return
	}

	// transform crd to our request object
	req := w.CRDToRequest(crd)

	// generate all the credentials
	creds, err := legacyGenCredentials(req, w.DBWorkerConfig)
	if err != nil {
		w.Error(fmt.Sprintf("unable to generate credentials err: %v", err))
		return
	}

	// store the credentials before creation just in case something breaks
	// store only the master secret at this point
	err = core.StoreDBCredentials(w.CredentialsStorer, &database.Credentials{database.CredTypeAdmin: creds[0]})
	if err != nil {
		w.Error(fmt.Sprintf("unable to store master credentials in kube-system err: %v", err))
		return
	}

	// create database for the request
	db, err := core.CreateDatabaseIfNotExist(w.DBCreateGetter, req, creds[database.CredTypeAdmin])
	if nil != err {
		w.Error(fmt.Sprintf("unable to create database err: %v", err))
		updateCRDStatus(w.StatusUpdater, w.Logger, crd.Name, s, database.StatusErrored, nil)
		return
	}
	updateCRDStatus(w.StatusUpdater, w.Logger, crd.Name, s, database.StatusUnavailable, &db.ID)

	// check and wait for DB to be available
	db, err = core.WaitForDBToBeAvailable(w.DBCreateGetter, db.ID, w.checkIntervalInMillis)
	if err != nil {
		w.Error(fmt.Sprintf("unable to get database status err: %v", err))
		return
	}

	updateCRDStatus(w.StatusUpdater, w.Logger, crd.Name, s, database.StatusAvailable, &db.ID)

	// enrich credentials with database info
	updatedCreds := addHostInfoToCredentials(creds, db)

	// store updated credentials
	err = core.StoreDBCredentials(w.CredentialsStorer, &updatedCreds)
	if err != nil {
		w.Error(fmt.Sprintf("unable to store credentials err: %v", err))
		return
	}

	// create metrics exporter
	err = core.CreateMetricsExporterForDB(w, getScope(req.Owner, w.DBWorkerConfig.nsSuffix), req.Name, creds[database.CredTypeMonitoring].ID)
	if err != nil {
		w.Error(fmt.Sprintf("unable to create metrics exporter err: %v", err))
		return
	}

}

// OnUpdate handles update event of postgresdb
func (w *DBWorker) OnUpdate(obj interface{}, newObj interface{}) {
	// TODO: fix no op
}

// OnDelete handles delete event of postgresdb
func (w *DBWorker) OnDelete(obj interface{}) {
	// TODO: fix no op
}

func updateCRDStatus(i core.StatusUpdater, l Logger, n string, s database.Scope, status database.Status, id *database.DatabaseID) {
	sReq := &database.StatusRequest{
		Name:   n,
		Status: status,
		ID:     id,
		Scope:  s,
	}
	err := core.UpdateStatus(i, sReq)
	if err != nil {
		l.Error(fmt.Sprintf("unable to update crd status, %v", err))
	}
}

func genCredentials(req *database.Request, c *DBWorkerConfig) (database.Credentials, error) {
	creds := make(database.Credentials)

	for k, credType := range database.GetAllCredentialTypes() {

		// generate password
		pw, err := core.GenPasswords(30)
		if err != nil {
			return nil, err
		}
		id := fmt.Sprintf("%s-%s-%s", req.Owner, req.Name, database.GetUserNameForType(credType))

		credential := &database.Credential{
			Password:     database.Password(*pw),
			Username:     database.GetUserNameForType(credType),
			CredType:     database.CredentialType(k),
			ID:           database.CredentialID(id),
			Scope:        getScopeForCredType(req.Owner, c.nsSuffix, credType),
			DatabaseName: "postgres",
		}
		creds[credType] = credential
	}
	return creds, nil
}

// DEPRECATED
func legacyGenCredentials(req *database.Request, c *DBWorkerConfig) (database.Credentials, error) {
	creds := make(database.Credentials)

	pw, err := core.GenPasswords(30)
	if err != nil {
		return nil, err
	}
	u := "master"

	for k, credType := range database.GetAllCredentialTypes() {

		id := fmt.Sprintf("%s-%s-%s", req.Owner, req.Name, database.GetUserNameForType(credType))

		credential := &database.Credential{
			Password:     database.Password(*pw),
			Username:     u,
			CredType:     database.CredentialType(k),
			ID:           database.CredentialID(id),
			Scope:        getScopeForCredType(req.Owner, c.nsSuffix, credType),
			DatabaseName: "postgres",
		}
		creds[credType] = credential
	}

	return creds, nil
}

func addHostInfoToCredentials(creds database.Credentials, db *database.Database) database.Credentials {
	updatedCreds := make(database.Credentials)
	for k, v := range creds {
		v.Host = db.Host
		v.Port = db.Port
		updatedCreds[k] = v
	}
	fmt.Println("In addHostInfoToCredentials")
	return updatedCreds
}

func getScopeForCredType(requestNamespace string, nsSuffix string, t database.CredentialType) database.Scope {
	retNS := requestNamespace
	if t == database.CredTypeAdmin {
		return database.Scope("kube-system")
	} else if t == database.CredTypeMonitoring {
		return getScope(requestNamespace, nsSuffix)
	}
	return database.Scope(retNS)
}

func getScope(ns, suffix string) database.Scope {
	if suffix != "" {
		return database.Scope(fmt.Sprintf("%s-%s", ns, suffix))
	}
	return database.Scope(ns)
}
