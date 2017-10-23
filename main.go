package main

import (
	"flag"
	"log"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/controller"

	dbClientSet "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned"
	dbInformer "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/informers/externalversions"
)

var version = "snapshot"

func main() {
	log.Printf("rds-controller version: %v", version)

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
		log.Printf("using kubeconfig %v", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		log.Printf("running inside cluster")
		config, err = rest.InClusterConfig()
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("error creating kubernetes client: %v", err)
	}

	dbClient, err := dbClientSet.NewForConfig(config)
	if err != nil {
		log.Fatalf("error creating db client: %v", err)
	}

	// dbInformerFactory acts like a cache for db resources like above
	dbInformerFactory := dbInformer.NewSharedInformerFactory(dbClient, 10*time.Minute)

	// this controller will deal with RDS dbs
	rdsController, err := controller.New(client, dbClient, dbInformerFactory)
	if err != nil {
		log.Fatalf("error creating db controller: %v", err)
	}

	// start go routines with our informers
	go dbInformerFactory.Start(nil)

	if err = rdsController.Run(2, nil); err != nil {
		log.Fatalf("Error running controller: %v", err)
	}
}
