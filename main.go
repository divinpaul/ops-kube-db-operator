package main

import (
	"flag"
	"log"
	"os"
	"time"

	"k8s.io/client-go/util/workqueue"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/gugahoi/rds-operator/pkg/controller"
)

func main() {
	version := "0.0.1"
	log.Printf("rds-controller version: %v", version)

	// read kube config file from flag
	var kubeconfig string
	flag.StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig file")
	flag.Parse()

	// if no flag has been passed, read kubeconfig file from environment
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	var err error
	var config *rest.Config
	// if flag has not been passed and env not set, presume running in cluster
	if kubeconfig != "" {
		log.Printf("using kubeconfig %v", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		log.Printf("running inside cluster")
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		log.Fatalf("error creating client: %v", err)
	}

	client := kubernetes.NewForConfigOrDie(config)

	// create the work queue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// sharedInformer acts like a cache for resources so that we dont hammer the api server
	sharedInformers := informers.NewSharedInformerFactory(client, 10*time.Minute)

	// this controller will deal with RDS dbs
	rdsController := controller.New(queue, client, sharedInformers.Core().V1().ConfigMaps())

	sharedInformers.Start(nil)
	rdsController.Run(2, nil)
}
