package main

import (
	"flag"
	"log"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	version := "0.0.0"
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
		log.Println("running inside cluster")
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		log.Fatalf("error creating client: %v", err)
	}

	client := kubernetes.NewForConfigOrDie(config)

	nsList, err := client.Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("unable to get namespaces: %v", err)
	}

	for _, ns := range nsList.Items {
		log.Printf("NAMESPACE: %v (%v): created on %v",
			ns.GetName(),
			ns.UID,
			ns.CreationTimestamp,
		)
	}
}
