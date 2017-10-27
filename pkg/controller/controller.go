package controller

import (
	"fmt"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/util/workqueue"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	dbClientSet "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned"
	dbInformerFactory "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/informers/externalversions"
	dbLister "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/listers/postgresdb/v1alpha1"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/pgdb"
)

// PgController is a controller for Postgres RDS DBs.
type PgController struct {
	// client is the standart kubernetes clientset
	client *kubernetes.Clientset

	// dbClient is the client for the DB crd
	dbClient dbClientSet.Interface

	// queue is where incoming work is placed - it handles de-dup and rate limiting
	queue workqueue.RateLimitingInterface

	// dbLister is the cache of DBs used for lookup
	lister dbLister.PostgresDBLister
	// dbSynced is the indicator of wether the cache is synced
	synced cache.InformerSynced

	// pg is how we interact with PostgresDB objects
	pgmgr *pgdb.Manager
}

// New instantiates an pgController
func New(
	client *kubernetes.Clientset,
	dbClient dbClientSet.Interface,
	dbInformer dbInformerFactory.SharedInformerFactory,
) (*PgController, error) {

	informer := dbInformer.Postgresdb().V1alpha1().PostgresDBs()

	lister := informer.Lister()
	pgmgr, err := pgdb.NewManager(client, dbClient, lister)
	if err != nil {
		return nil, err
	}
	c := &PgController{
		client:   client,
		dbClient: dbClient,
		lister:   lister,
		synced:   informer.Informer().HasSynced,
		queue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "DB"),
		pgmgr:    pgmgr,
	}

	log.Info("Setting up event handlers")
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueue,
		UpdateFunc: func(old, new interface{}) {
			if !reflect.DeepEqual(old, new) {
				c.enqueue(new)
			}
		},
		DeleteFunc: c.enqueue,
	})

	return c, nil
}

// Run starts the controller
func (p *PgController) Run(threadiness int, stopChan <-chan struct{}) error {
	// do not allow panics to crash the controller
	defer runtime.HandleCrash()

	// shutdown the queue when done
	defer p.queue.ShutDown()

	log.Info("Starting Pg Controller")

	log.Info("waiting for cache to sync")
	if !cache.WaitForCacheSync(stopChan, p.synced) {
		return fmt.Errorf("timeout waiting for sync")
	}
	log.Info("caches synced successfully")

	for i := 0; i < threadiness; i++ {
		go wait.Until(p.runWorker, time.Second, stopChan)
	}

	// block until we are told to exit
	<-stopChan
	return nil
}

func (p *PgController) runWorker() {
	// process the next item in queue until it is empty
	for p.processNextWorkItem() {
	}
}

func (p *PgController) processNextWorkItem() bool {
	// get next item from work queue
	key, quit := p.queue.Get()
	if quit {
		return false
	}

	// indicate to queue when work is finished on a specific item
	defer p.queue.Done(key)

	//  Sync is to push changes for a postgresdb resource
	err := p.pgmgr.Sync(key.(string))
	if err == nil {
		// processed successfully, lets forget item in queue and return success
		p.queue.Forget(key)
		return true
	}

	// There was an error processing the item, log and requeue
	runtime.HandleError(err)

	// Add item back in with a rate limited backoff
	p.queue.AddRateLimited(key)

	return true
}

func (p *PgController) enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(fmt.Errorf("error obtaining key for enqueued object: %v", err))
	}
	log.Infof("Enqueueing: %s", key)
	p.queue.Add(key)
}
