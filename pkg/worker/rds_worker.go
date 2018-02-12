package worker

import (
	"fmt"

	"github.com/golang/glog"

	crds "github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	postgresdbv1alpha1 "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned/typed/postgresdb/v1alpha1"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/k8s"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/postgres"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/rds"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/secret"
	"k8s.io/client-go/kubernetes"
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
	dbInstanceCreator rds.DBInstanceCreator
	k8sClient         *k8s.Client
	crdClient         k8s.CRDClient

	// env config
	config *RDSConfig
}

// OnCreate handles create event for postgresDB crd and
// will orchestrate creation of RDS instance, db credentials in secrets
// and creation of metrics exporter deployment
func (w *RDSWorker) OnCreate(obj interface{}) {
	crd := obj.(*crds.PostgresDB)
	crdName := crd.ObjectMeta.Name
	crdNamespace := crd.ObjectMeta.Namespace
	instanceName := fmt.Sprintf("%s-%s", crdName, crd.GetUID())

	masterUser, _ := postgres.NewMasterUser()
	masterScrt, err := w.k8sClient.SaveMasterSecret(crdName, masterUser, nil, instanceName)

	crd.Status.Ready = fmt.Sprintf("Creating database instance %s...", instanceName)
	crd, err = w.crdClient.Update(crd)
	if err != nil {
		glog.Errorf("There was an error updating the crd: %s", err.Error())
		return
	}

	createdInstance, err := w.createInstance(crd, masterScrt, instanceName)

	if err != nil {
		glog.Errorf("There was an error creating database instance %s: %s", instanceName, err.Error())
		crd.Status.Ready = fmt.Sprintf("There was an error creating database instance %s: %s", instanceName, err.Error())
		crd, err = w.crdClient.Update(crd)
		if err != nil {
			glog.Errorf("There was an error updating the crd: %s", err.Error())
		}

		return
	}

	crd.Status.Ready = "available"
	crd.Status.ARN = createdInstance.ARN
	crd, err = w.crdClient.Update(crd)
	if err != nil {
		glog.Errorf("There was an error updating the crd: %s", err.Error())
		return
	}

	if createdInstance.AlreadyExists {
		glog.Infof("Database instance %s already exists, so finishing...", instanceName)
		return
	}

	w.k8sClient.SaveMasterSecret(crdName, masterUser, createdInstance, instanceName)

	// TODO: use postgres package to create new database and db users
	conn, _ := postgres.NewRawConn(createdInstance.Address, createdInstance.Port, masterUser)
	dd, _ := conn.CreateDatabaseDescriptor()

	w.k8sClient.SaveAdminSecret(crd, dd, instanceName)
	w.k8sClient.SaveMetricsExporterSecret(crd, dd, instanceName)
	metricsExporter := postgres.NewMetricsExporter(w.k8sClient.Clientset)
	err = metricsExporter.Deploy(fmt.Sprintf("%s-shadow", crdNamespace), crdName)
	if err != nil {
		glog.Errorf("There was an error creating exporter-deployment %s: %s", crdName, err.Error())
	}
	// TODO: Create RDS Cloudwatch metrics exporter (TBD) deployment in shadow ns
}

// OnUpdate handles update event of postgresdb
func (w *RDSWorker) OnUpdate(obj interface{}, newObj interface{}) {
	// TODO: fix no op
}

// OnDelete handles delete event of postgresdb
func (w *RDSWorker) OnDelete(obj interface{}) {
	// TODO: fix no op
}

// NewRDSWorker returns new RDSWorker instance for handling change events on postgresDB crd
func NewRDSWorker(dbInstanceCreator rds.DBInstanceCreator, clientSet kubernetes.Interface, crdClientset postgresdbv1alpha1.PostgresdbV1alpha1Interface, config *RDSConfig, crdClient k8s.CRDClient) *RDSWorker {
	k8sClient := k8s.NewClient(clientSet, crdClientset)

	return &RDSWorker{
		dbInstanceCreator: dbInstanceCreator,
		k8sClient:         k8sClient,
		config:            config,
		crdClient:         crdClient,
	}
}

func (w *RDSWorker) createInstance(crd *crds.PostgresDB, masterScrt *secret.DBSecret, instanceName string) (*rds.CreateInstanceOutput, error) {
	glog.Infof("Creating database %s...", instanceName)
	createdDB, err := w.dbInstanceCreator.Create(&rds.CreateInstanceInput{
		InstanceName:   instanceName,
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

	return createdDB, err
}
