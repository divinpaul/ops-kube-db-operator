package controller

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/util/workqueue"

	"k8s.io/apimachinery/pkg/util/runtime"
	informercorev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	listercorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// RDSController is a controller for RDS DBs.
type RDSController struct {
	// cmGetter is a configMap getter
	cmGetter corev1.ConfigMapsGetter
	// cmLister is a secondary cache of configMaps used for lookups
	cmLister listercorev1.ConfigMapLister
	// cmSynces is a flag to indicate if the cache is synced
	cmSynced cache.InformerSynced
	// queue is where incoming work is placed - it handles de-dup and rate limiting
	queue workqueue.RateLimitingInterface
}

// New instantiates an rds controller
func New(
	queue workqueue.RateLimitingInterface,
	client *kubernetes.Clientset,
	cmInformer informercorev1.ConfigMapInformer,
) *RDSController {

	c := &RDSController{
		cmGetter: client.CoreV1(),
		cmLister: cmInformer.Lister(),
		cmSynced: cmInformer.Informer().HasSynced,
		queue:    queue,
	}

	cmInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: c.enqueue,
			UpdateFunc: func(old, new interface{}) {
				if !reflect.DeepEqual(old, new) {
					c.enqueue(new)
				}
			},
			DeleteFunc: c.enqueue,
		},
	)

	return c
}

// Run starts the controller
func (c *RDSController) Run(threadiness int, stopChan <-chan struct{}) {
	// do not allow panics to crash the controller
	defer runtime.HandleCrash()

	// shutdown the queue when done
	defer c.queue.ShutDown()

	log.Print("Starting RDS Controller")

	log.Print("waiting for cache to sync")
	if !cache.WaitForCacheSync(stopChan, c.cmSynced) {
		log.Print("timeout waiting for sync")
		return
	}
	log.Print("caches synced successfully")

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopChan)
	}

	// block until we are told to exit
	<-stopChan
}

func (c *RDSController) runWorker() {
	// process the next item in queue until it is empty
	for c.processNextWorkItem() {
	}
}

func (c *RDSController) processNextWorkItem() bool {
	// get next item from work queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	// indicate to queue when work is finished on a specific item
	defer c.queue.Done(key)

	log.Printf("DO SOME STUFF: %v", key)
	return true

	// TODO: actually do something
	// err := c.DoSomething(key.(string))
	// if err == nil {
	// // processed succesfully, lets forget item in queue and return success
	// 	c.queue.Forget(key)
	// 	return true
	// }

	// There was an error processing the item, log and requeue
	// runtime.HandleError("some error", err)

	// Add item back in with a rate limited backoff
	// c.queue.AddRateLimited(key)

	// return true
}

func (c *RDSController) enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(fmt.Errorf("error obtaining key for enqueued object: %v", err))
	}
	c.queue.Add(key)
}
