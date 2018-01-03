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

// RDSWorker creates an RDS instance for every postres
// DB requested. It containes all the config that will
// change per env.
type RDSWorker struct {
	// injected deps for testing
	rds          db.RDSManager
	clientset    kubernetes.Interface
	crdClientset postgresdbv1alpha1.PostgresdbV1alpha1Interface

	// env config
	backupRetentionPeriod int64
	multiAZ               bool
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

	// tags := []*db.Tag{
	// 	&db.Tag{
	// 		Key:   &string("Namespace"),
	// 		Value: crdNamespace,
	// 	},
	// 	&db.Tag{
	// 		Key:   "Resource",
	// 		Value: crdName,
	// 	},
	// 	&db.Tag{
	// 		Key:   "CreatedBy",
	// 		Value: "ops-kube-db-operator",
	// 	},
	// }

	// TODO: make this env passed in
	createdDB, err := w.rds.Create(&db.DB{
		Name:               &dbName,
		MasterUsername:     &masterScrt.Username,
		MasterUserPassword: &masterScrt.Password,
		// Tags:               tags,
	}, db.ProductionDefaults)

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

func NewRDSWorker(rds db.RDSManager, clientSet kubernetes.Interface, crdClientset postgresdbv1alpha1.PostgresdbV1alpha1Interface) *RDSWorker {
	return &RDSWorker{
		rds:          rds,
		clientset:    clientSet,
		crdClientset: crdClientset,
	}
}
