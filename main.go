package main

import (
	"flag"
	"os"
	"strings"

	"github.com/golang/glog"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/controller"

	"time"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/informers/externalversions"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/k8s"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/rds"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/signals"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/worker"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	rds2 "github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
)

var kubeconfig string
var region string
var subnetGroup string
var sgIDs []*string
var nsSuffix string

func main() {

	if subnetGroup == "" {
		glog.Fatalf("please provide a subnet group")
	}

	if len(sgIDs) == 0 {
		glog.Fatalf("please provide a comma separated list of security group ids with at least one id.")
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	var config *rest.Config
	var err error

	// if flag has not been passed and env not set, presume running in cluster
	if kubeconfig != "" {
		glog.Infof("using kubeconfig %v", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		glog.Infof("running inside cluster")
		config, err = rest.InClusterConfig()
	}

	if nil != err {
		glog.Errorf("error creating config, %v", err)
		return
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("error building k8s clientset: %s", err.Error())
	}

	crdClient, err := clientset.NewForConfig(config)
	if err != nil {
		glog.Fatalf("error building CRD clientset: %s", err.Error())
	}

	rdsClient, err := getRDSClient()
	if err != nil {
		glog.Fatalf("error cannot get rds client: %s", err.Error())
	}

	rdsConfig := rds.NewRDSTransformerConfig(&subnetGroup, sgIDs)
	rdsTransformer := rds.NewBumblebee(rdsConfig)
	wrkr := worker.NewDBWorker(
		rds.NewRDSImpure(rdsClient, rdsTransformer),
		k8s.NewStoreCreds(k8sClient),
		k8s.NewMetricsExporter(k8sClient),
		worker.NewConfig(100000, nsSuffix),
		worker.NewPostgresDBValidator(),
		worker.NewLogger(),
		worker.NewOptimus(),
		k8s.NewCRDClient(crdClient),
	)

	factory := externalversions.NewSharedInformerFactory(crdClient, time.Second*30)
	go factory.Start(stopCh)

	crdController := controller.New(factory, wrkr)
	crdController.Run(stopCh)
}

func getRDSClient() (rdsiface.RDSAPI, error) {
	c := aws.NewConfig().WithRegion(region)
	s, err := session.NewSession(c)
	if err != nil {
		return nil, err
	}
	return rds2.New(s), nil
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig file")
	flag.StringVar(&nsSuffix, "ns-suffix", "", "namespace suffix (or env NS_SUFFIX)")
	flag.Parse()

	// if no flag has been passed, read kubeconfig file from environment
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}
	region = os.Getenv("AWS_REGION")
	subnetGroup = os.Getenv("DB_SUBNET_GROUP")
	sgList := os.Getenv("DB_SECURITY_GROUP_IDS")

	if region == "" {
		region = "ap-southeast-2"
	}

	t := strings.Split(sgList, ",")
	for _, v := range t {
		sgIDs = append(sgIDs, &v)
	}

	if nsSuffix == "" {
		nsSuffix = os.Getenv("NS_SUFFIX")
	}
}
