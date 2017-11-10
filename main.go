package main

import (
	"flag"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/controller"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/signals"

	clientset "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned"
)

var version = "snapshot"

func main() {

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	// read kube config file from flag
	var kubeconfig string
	flag.StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig file")
	flag.Parse()

	// if no flag has been passed, read kubeconfig file from environment
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	var config *rest.Config
	var err error
	// if flag has not been passed and env not set, presume running in cluster
	if kubeconfig != "" {
		log.Infof("using kubeconfig %v", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		log.Infof("running inside cluster")
		config, err = rest.InClusterConfig()
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("error creating kubernetes client: %v", err)
	}

	dbClient, err := clientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("error creating db client: %v", err)
	}

	// this controller will deal with RDS dbs
	rdsController, err := controller.New(kubeClient, dbClient, stopCh)
	if err != nil {
		log.Fatalf("error creating db controller: %v", err)
	}

	if err = rdsController.Run(2); err != nil {
		log.Fatalf("error running controller: %v", err)
	}

}

func init() {
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	loglevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if loglevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else if loglevel == "warn" {
		log.SetLevel(log.WarnLevel)
	} else if loglevel == "error" {
		log.SetLevel(log.ErrorLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.Infof("postgresdb-controller version: %v", version)
}
