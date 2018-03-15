package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned"
	informers "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/informers/externalversions"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/controller"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/k8s"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/rds"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/signals"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/worker"
)

var version = "snapshot"
var dbEnvironment string
var kubeconfig string
var subnetGroup string
var sgList string
var sgIDs []*string

func main() {
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	var kubeconfig string

	var config *rest.Config
	var err error

	// if flag has not been passed and env not set, presume running in cluster
	if kubeconfig != "" {
		glog.Infof("using kubeconfig %v", kubeconfig)
		config, _ = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		glog.Infof("running inside cluster")
		config, _ = rest.InClusterConfig()
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Error building k8s clientset: %s", err.Error())
	}

	dbClient, err := clientset.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Error building CRD clientset: %s", err.Error())
	}

	backups := false
	multiAZ := false
	if dbEnvironment == "production" {
		backups = true
		multiAZ = true
	}

	manager := rds.NewDBInstanceManager()
	rdsConfig := &worker.RDSConfig{
		OperatorVersion: version,
		DefaultSize:     "t1.Small", // hardcoded for now
		DefaultStorage:  5,          // hardcoded for now
		DBEnvironment:   dbEnvironment,
		BackupRetention: backups,
		MultiAZ:         multiAZ,
		SubnetGroup:     subnetGroup,
		SecurityGroups:  sgIDs,
	}
	rdsWorker := worker.NewRDSWorker(manager, k8sClient, dbClient.PostgresdbV1alpha1(), rdsConfig, k8s.NewK8SCRDClient(dbClient.PostgresdbV1alpha1()))

	dbInformerFactory := informers.NewSharedInformerFactory(dbClient, time.Second*30)
	go dbInformerFactory.Start(stopCh)

	rdsController := controller.New(dbInformerFactory, rdsWorker)
	rdsController.Run(stopCh)
}

func init() {
	flag.StringVar(&dbEnvironment, "dbenv", "development", "Environment for creating RDS instances (production|development).")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig file")
	flag.Parse()

	// if no flag has been passed, read kubeconfig file from environment
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	subnetGroup = os.Getenv("DB_SUBNET_GROUP")
	sgList = os.Getenv("DB_SECURITY_GROUP_IDS")

	if subnetGroup == "" {
		glog.Error("please provide a subnet group")
		return
	}

	if sgList == "" {
		glog.Error("please provide a comma separated list of security group ids.")
		return
	}

	t := strings.Split(sgList, ",")
	for _, v := range t {
		sgIDs = append(sgIDs, &v)
	}
}
