package worker

import (
	"fmt"
	"time"

	crds "github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	postgresdbv1alpha1 "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned/typed/postgresdb/v1alpha1"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/db"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/postgres"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/secret"
	"k8s.io/client-go/kubernetes"
)

type RDSConfig struct {
	DefaultSize     string
	DefaultStorage  int64
	DBEnvironment   string
	OperatorVersion string
}

// RDSWorker creates an RDS instance for every postres
// DB requested. It containes all the config that will
// change per env.
type RDSWorker struct {
	// injected deps for testing
	rds          db.RDSManager
	clientset    kubernetes.Interface
	crdClientset postgresdbv1alpha1.PostgresdbV1alpha1Interface

	// env config
	config *RDSConfig
}

func (w *RDSWorker) OnCreate(obj interface{}) {
	crd := obj.(*crds.PostgresDB)
	crdName := crd.ObjectMeta.Name
	crdNamespace := crd.ObjectMeta.Namespace
	dbName := fmt.Sprintf("k8s-%s-%s", crdNamespace, crdName)
	masterSecretName := fmt.Sprintf("%s-master", dbName)
	adminSecretName := fmt.Sprintf("%s-admin", dbName)

	// Create and save master password secret first
	_, masterScrt, _ := secret.NewOrGet(w.clientset.CoreV1(), "kube-system", masterSecretName)
	creds, _ := postgres.GenPasswords(2, 20)
	masterScrt.Username = creds[0]
	masterScrt.Password = creds[1]
	secret.SaveOrCreate(w.clientset.CoreV1(), masterScrt)

	tags := make([]*db.Tag, 5)

	tags = append(tags, tag("Namespace", crdNamespace))
	tags = append(tags, tag("Resource", crdName))
	tags = append(tags, tag("DBEnvironment", w.config.DBEnvironment))
	tags = append(tags, tag("CreatedBy", "ops-kube-db-operator"))
	tags = append(tags, tag("OperatorVersion", w.config.OperatorVersion))

	// set env defaults for database
	dbDefaults := db.DevelopmentDefaults
	if w.config.DBEnvironment == "production" {
		dbDefaults = db.ProductionDefaults
	}

	dbSize := w.config.DefaultSize
	if crd.Spec.Size != "" {
		dbSize = crd.Spec.Size
	}

	dbStorage := w.config.DefaultStorage
	if crd.Spec.Storage != 0 {
		dbStorage = crd.Spec.Storage
	}

	createdDB, err := w.rds.Create(&db.DB{
		Name:               &dbName,
		MasterUsername:     &masterScrt.Username,
		MasterUserPassword: &masterScrt.Password,
		DBInstanceClass:    &dbSize,
		StorageAllocatedGB: &dbStorage,
		Tags:               tags,
	}, dbDefaults)

	if err != nil {
		crd.Status.Ready = fmt.Sprintf("Error creating Database: %s", err.Error())
		w.crdClientset.PostgresDBs(crdNamespace).Update(crd)
		return
	}

	crd.Status.ARN = *createdDB.ARN
	w.crdClientset.PostgresDBs(crdNamespace).Update(crd)

	// wait for DB to become ready and timeout after 20 mins
	timeoutCount := 0
	var statDB *db.DB
	for {
		statDB, _ = w.rds.Stat(dbName)
		crd.Status.Ready = *statDB.Status
		w.crdClientset.PostgresDBs(crdNamespace).Update(crd)

		if *statDB.Status == db.StatusAvailable {
			break
		}

		timeoutCount++
		if timeoutCount > 40 {
			crd.Status.Ready = "Timed out while waiting for the DB to become available."
			w.crdClientset.PostgresDBs(crdNamespace).Update(crd)
			return
		}

		time.Sleep(30 * time.Second)
	}

	// update master secret
	masterScrt.Host = *statDB.Address
	masterScrt.Port = string(*statDB.Port)
	masterScrt.DatabaseName = "postgres"
	secret.SaveOrCreate(w.clientset.CoreV1(), masterScrt)

	// TODO: use postgres package to create new database and db users

	// create db users in crd namespace
	_, adminScrt, _ := secret.NewOrGet(w.clientset.CoreV1(), crdNamespace, adminSecretName)
	adminScrt.Username = masterScrt.Username
	adminScrt.Password = masterScrt.Password
	adminScrt.Host = *statDB.Address
	adminScrt.Port = string(*statDB.Port)
	adminScrt.DatabaseName = "postgres"
	secret.SaveOrCreate(w.clientset.CoreV1(), adminScrt)
}

func (w *RDSWorker) OnUpdate(obj interface{}, newObj interface{}) {
	// TODO: fix no op
}

func (w *RDSWorker) OnDelete(obj interface{}) {
	// TODO: fix no op
}

func NewRDSWorker(rds db.RDSManager, clientSet kubernetes.Interface, crdClientset postgresdbv1alpha1.PostgresdbV1alpha1Interface, config *RDSConfig) *RDSWorker {
	return &RDSWorker{
		rds:          rds,
		clientset:    clientSet,
		crdClientset: crdClientset,
		config:       config,
	}
}

func tag(key, val string) *db.Tag {
	return &db.Tag{
		Key:   &key,
		Value: &val,
	}
}
