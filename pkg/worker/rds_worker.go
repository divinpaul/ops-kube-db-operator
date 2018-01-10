package worker

import (
	"fmt"

	"github.com/golang/glog"

	crds "github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	postgresdbv1alpha1 "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned/typed/postgresdb/v1alpha1"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/postgres"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/rds"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/secret"
	"k8s.io/client-go/kubernetes"
)

const (
	OperatorAdminNamespace = "kube-system"
)

type RDSConfig struct {
	DefaultSize     string
	DefaultStorage  int64
	DBEnvironment   string
	OperatorVersion string
}

// RDSWorker creates an RDS instance for every postgres
// DB requested. It contains all the config that will
// change per env.
type RDSWorker struct {
	// injected deps for testing
	dbInstanceCreator          rds.DBInstanceCreator
	clientset    kubernetes.Interface
	crdClientset postgresdbv1alpha1.PostgresdbV1alpha1Interface

	// env config
	config *RDSConfig
}

func (w *RDSWorker) OnCreate(obj interface{}) {
	crd := obj.(*crds.PostgresDB)
	dbName := fmt.Sprintf("%s-%s", crd.ObjectMeta.Name, crd.GetUID())
	masterSecretName := fmt.Sprintf("%s-master", crd.ObjectMeta.Name)
	adminSecretName := fmt.Sprintf("%s-admin", crd.ObjectMeta.Name)

	// Create and save master password secret first
	glog.Infof("Creating master secret for database %s...", dbName)
	masterScrt, _ := secret.NewOrGet(w.clientset.CoreV1(), OperatorAdminNamespace, masterSecretName)
	creds, _ := postgres.GenPasswords(1, 30)
	masterScrt.Username = "master"
	masterScrt.Password = creds[0]

	crd.Status.Ready = fmt.Sprintf("Creating database %s...", dbName)
	w.crdClientset.PostgresDBs(crd.ObjectMeta.Namespace).Update(crd)

	glog.Infof("Creating database %s...", dbName)
	createdDB, err := w.dbInstanceCreator.Create(&rds.CreateInstanceInput{
		InstanceName:   dbName,
		Storage:        crd.Spec.Storage,
		Size:           crd.Spec.Size,
		MasterPassword: masterScrt.Password,
		MasterUsername: masterScrt.Username,
		Backups:        false,
		MultiAZ:        false,
		Tags: map[string]string{
			"Namespace":                crd.ObjectMeta.Namespace,
			"Resource":                 crd.ObjectMeta.Name,
			"DBEnvironment":            w.config.DBEnvironment,
			"CreatedBy":                "ops-kube-db-operator",
			"CreatedByOperatorVersion": w.config.OperatorVersion,
		},
	})

	if err != nil {
		glog.Errorf("There was an error creating database %s: %s", dbName, err.Error())
		crd.Status.Ready = fmt.Sprintf("Error creating Database: %s", err.Error())
		w.crdClientset.PostgresDBs(crd.ObjectMeta.Namespace).Update(crd)
		return
	}

	crd.Status.Ready = "available"
	crd.Status.ARN = createdDB.ARN
	w.crdClientset.PostgresDBs(crd.ObjectMeta.Namespace).Update(crd)

	if createdDB.AlreadyExists {
		glog.Infof("Database %s already exists, so finishing...", dbName)
		return
	}

	// update master secret
	glog.Infof("Updating master secret for database %s with address...", dbName)
	masterScrt.Host = createdDB.Address
	masterScrt.Port = string(createdDB.Port)
	masterScrt.DatabaseName = "postgres"
	glog.Infof("Saving secret %s...", masterScrt)
	masterScrt.Save()

	// TODO: use postgres package to create new database and db users

	// create db secrets in crd namespace
	glog.Infof("Creating admin secret for database %s...", dbName)
	adminScrt, _ := secret.NewOrGet(w.clientset.CoreV1(), crd.ObjectMeta.Namespace, adminSecretName)
	adminScrt.Username = masterScrt.Username
	adminScrt.Password = masterScrt.Password
	adminScrt.Host = masterScrt.Host
	adminScrt.Port = masterScrt.Port
	adminScrt.DatabaseName = "postgres"
	glog.Infof("Saving secret %s...", adminScrt)
	adminScrt.Save()

	// TODO: Create secret in shadow ns for exporter

	// TODO: Create postgres exporter deployment in shadow ns

	// TODO: Create RDS Cloudwatch metrics exporter (TBD) deployment in shadow ns
}

func (w *RDSWorker) OnUpdate(obj interface{}, newObj interface{}) {
	// TODO: fix no op
}

func (w *RDSWorker) OnDelete(obj interface{}) {
	// TODO: fix no op
}

func NewRDSWorker(dbInstanceCreator rds.DBInstanceCreator, clientSet kubernetes.Interface, crdClientset postgresdbv1alpha1.PostgresdbV1alpha1Interface, config *RDSConfig) *RDSWorker {
	return &RDSWorker{
		dbInstanceCreator:          dbInstanceCreator,
		clientset:    clientSet,
		crdClientset: crdClientset,
		config:       config,
	}
}
