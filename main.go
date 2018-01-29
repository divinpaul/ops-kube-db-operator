package main

import (
	"flag"
	"os"
	"time"

	"github.com/golang/glog"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned"
	informers "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/informers/externalversions"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/controller"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/rds"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/signals"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/worker"
)

var (
	version = "snapshot"
	dbEnvironment string
	kubeconfig string
)

func main() {
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

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

	manager := rds.NewDBInstanceManager()
	rdsConfig := &worker.RDSConfig{
		OperatorVersion: version,
		DefaultSize:     "t1.Small", // hardcoded for now
		DefaultStorage:  5,          // hardcoded for now
		DBEnvironment:   dbEnvironment,
	}
	rdsWorker := worker.NewRDSWorker(manager, k8sClient, dbClient.PostgresdbV1alpha1(), rdsConfig)

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
}
