package controller

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/util/workqueue"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	dbClientSet "github.com/gugahoi/rds-operator/pkg/client/clientset/versioned"
	dbInformerFactory "github.com/gugahoi/rds-operator/pkg/client/informers/externalversions"
	dbLister "github.com/gugahoi/rds-operator/pkg/client/listers/db/v1alpha1"

	"github.com/MYOB-Technology/dataform/pkg/db"
	"github.com/MYOB-Technology/dataform/pkg/service"
)

// RDSController is a controller for RDS DBs.
type RDSController struct {
	// client is the standart kubernetes clientset
	client kubernetes.Interface

	// dbClient is the client for the DB crd
	dbClient dbClientSet.Interface

	// queue is where incoming work is placed - it handles de-dup and rate limiting
	queue workqueue.RateLimitingInterface

	// dbLister is the cache of DBs used for lookup
	lister dbLister.DBLister
	// dbSynced is the indicator of wether the cache is synced
	synced cache.InformerSynced

	// rds is how we interact with AWS RDS Service
	rds *db.Manager
}

// New instantiates an rds controller
func New(
	client *kubernetes.Clientset,
	dbClient dbClientSet.Interface,
	dbInformer dbInformerFactory.SharedInformerFactory,
) *RDSController {

	informer := dbInformer.Db().V1alpha1().DBs()

	c := &RDSController{
		client:   client,
		dbClient: dbClient,
		lister:   informer.Lister(),
		synced:   informer.Informer().HasSynced,
		queue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "DB"),
		rds:      db.NewManager(service.New("")),
	}

	log.Print("Setting up event handlers")
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueue,
		UpdateFunc: func(old, new interface{}) {
			if !reflect.DeepEqual(old, new) {
				c.enqueue(new)
			}
		},
		DeleteFunc: c.enqueue,
	})

	return c
}

// Run starts the controller
func (c *RDSController) Run(threadiness int, stopChan <-chan struct{}) error {
	// do not allow panics to crash the controller
	defer runtime.HandleCrash()

	// shutdown the queue when done
	defer c.queue.ShutDown()

	log.Print("Starting RDS Controller")

	log.Print("waiting for cache to sync")
	if !cache.WaitForCacheSync(stopChan, c.synced) {
		return fmt.Errorf("timeout waiting for sync")
	}
	log.Print("caches synced successfully")

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopChan)
	}

	// block until we are told to exit
	<-stopChan
	return nil
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

	err := c.processDB(key.(string))
	if err == nil {
		// processed succesfully, lets forget item in queue and return success
		c.queue.Forget(key)
		return true
	}

	// There was an error processing the item, log and requeue
	runtime.HandleError(fmt.Errorf("%v", err))

	// Add item back in with a rate limited backoff
	c.queue.AddRateLimited(key)

	return true
}

func (c *RDSController) enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(fmt.Errorf("error obtaining key for enqueued object: %v", err))
	}
	c.queue.Add(key)
}

func (c *RDSController) processDB(key string) error {
	// get resource name and namespace out of key
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("error splitting namespace/key from obj %s: %v", key, err)
	}
	log.Printf("HEYOH: %s/%s", ns, name)

	db, err := c.lister.DBs(ns).Get(name)
	if err != nil {
		log.Printf("failed to retrieve up to date db resource %s, it has most likely been deleted: %v", key, err)
	}
	log.Printf("%s: %v", key, db.Spec.Type)

	log.Printf("Finished updating %s", key)
	return nil
}
