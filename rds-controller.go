package main

import (
	"log"

	"k8s.io/apimachinery/pkg/util/runtime"
	informercorev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	listercorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// RDSController is a controller for RDS DBs.
type RDSController struct {
	cmGetter corev1.ConfigMapsGetter
	cmLister listercorev1.ConfigMapLister
	cmSynced cache.InformerSynced
}

// NewRDSController instantiates an rds controller
func NewRDSController(
	client *kubernetes.Clientset,
	cmInformer informercorev1.ConfigMapInformer,
) *RDSController {

	c := &RDSController{
		cmGetter: client.CoreV1(),
		cmLister: cmInformer.Lister(),
		cmSynced: cmInformer.Informer().HasSynced,
	}

	cmInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				c.onAdd(obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				c.onUpdate(oldObj, newObj)
			},
			DeleteFunc: func(obj interface{}) {
				c.onDelete(obj)
			},
		},
	)

	return c
}

// Run starts the controller
func (c *RDSController) Run(stopChan <-chan struct{}) {
	log.Println("waiting for cache to sync")
	if !cache.WaitForCacheSync(stopChan, c.cmSynced) {
		log.Print("timeout waiting for sync")
		return
	}
	log.Println("caches synced successfully")
	<-stopChan
}

func (c *RDSController) onAdd(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		log.Printf("onAdd: error getting key for %#v: %v", obj, err)
		runtime.HandleError(err)
	}
	log.Printf("onAdd: %v", key)
}

func (c *RDSController) onUpdate(old, new interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(old)
	if err != nil {
		log.Printf("onUpdate: error getting key for %#v: %v", old, err)
		runtime.HandleError(err)
	}
	log.Printf("onUpdate: %v", key)
}

func (c *RDSController) onDelete(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		log.Printf("onDelete: error getting key for %#v: %v", obj, err)
		runtime.HandleError(err)
	}
	log.Printf("onDelete: %v", key)
}
