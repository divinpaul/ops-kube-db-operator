package controller

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	dbClientSet "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned"
	dbInformerFactory "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/informers/externalversions"
	dbLister "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/listers/postgresdb/v1alpha1"

	"github.com/MYOB-Technology/dataform/pkg/db"
	"github.com/MYOB-Technology/dataform/pkg/service"
)

// DBInstanceConfig holds all cloud provider specific config
type DBInstanceConfig struct {
	Port            int64
	SubnetGroupName string
	SecurityGroupId string
	Multizone       bool
	EncryptionKey   string
}

// RDSController is a controller for RDS DBs.
type RDSController struct {
	// client is the standart kubernetes clientset
	client kubernetes.Interface

	// dbClient is the client for the DB crd
	dbClient dbClientSet.Interface

	// queue is where incoming work is placed - it handles de-dup and rate limiting
	queue workqueue.RateLimitingInterface

	// dbLister is the cache of DBs used for lookup
	lister dbLister.PostgresDBLister
	// dbSynced is the indicator of wether the cache is synced
	synced cache.InformerSynced

	// rds is how we interact with AWS RDS Service
	rds *db.Manager

	// dbConfig defines cloud provider specifics
	dbConfig *DBInstanceConfig
}

// New instantiates an rds controller
func New(
	client *kubernetes.Clientset,
	dbClient dbClientSet.Interface,
	dbInformer dbInformerFactory.SharedInformerFactory,
	dbConfig *DBInstanceConfig,
) *RDSController {

	informer := dbInformer.Postgresdb().V1alpha1().PostgresDBs()

	c := &RDSController{
		client:   client,
		dbClient: dbClient,
		lister:   informer.Lister(),
		synced:   informer.Informer().HasSynced,
		queue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "DB"),
		rds:      db.NewManager(service.New("")),
		dbConfig: dbConfig,
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

func NewDBInstanceConfig() (*DBInstanceConfig, error) {

	dbSubnetGroupName := os.Getenv("DB_SUBNETGROUPNAME")
	dbSecurityGroupId := os.Getenv("DB_SECURITYGROUPID")
	dbEncryptionKeyArn := os.Getenv("DB_ENCRYPTIONKEYARN")
	dbPortString := os.Getenv("DB_PORT")

	// retrieve db port if provided
	var dbPort int64
	var err error
	if dbPortString == "" {
		dbPort = 5432
	} else {
		dbPort, err = strconv.ParseInt(dbPortString, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing port : %s", dbPortString)
		}
	}

	if dbSecurityGroupId == "" {
		return nil, fmt.Errorf("error: required environment variable DB_SECURITYGROUPID missing")
	}
	if dbSubnetGroupName == "" {
		return nil, fmt.Errorf("error: required environment variable DB_SUBNETGROUPNAME missing")
	}

	dbConfig := DBInstanceConfig{
		Port:            dbPort,
		SubnetGroupName: dbSubnetGroupName,
		SecurityGroupId: dbSecurityGroupId,
		Multizone:       true,
		EncryptionKey:   dbEncryptionKeyArn,
	}

	return &dbConfig, nil

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
	log.Printf("Enqueueing: %s", key)
	c.queue.Add(key)
}

func (c *RDSController) processDB(key string) error {
	log.Printf("Processing DB: %s", key)
	// get resource name and namespace out of key
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("error splitting namespace/key from obj %s: %v", key, err)
	}

	db, err := c.lister.PostgresDBs(ns).Get(name)
	if err != nil {
		log.Printf("failed to retrieve up to date db resource %s, it has most likely been deleted: %v", key, err)
		return nil
	}
	log.Printf("%s: %v", key, db.Spec.Type)

	// deep copy to not change the cache
	newDbInterface, _ := scheme.Scheme.DeepCopy(db)
	newDb := newDbInterface.(*v1alpha1.PostgresDB)

	// create the db now
	newObj, err := createDb(c.rds, c.dbConfig, newDb)
	if err != nil {
		// failed creating the database
		return fmt.Errorf("failed creating db %s: requeuing - %v", key, err)
	}
	// update the db information
	_, err = c.dbClient.Postgresdb().PostgresDBs(ns).Update(newObj)
	if err != nil {
		return fmt.Errorf("failed to update db %s: %v", key, err)
	}

	log.Printf("Finished updating %s", key)
	return nil
}

func createDb(rds *db.Manager, dbConfig *DBInstanceConfig, resource *v1alpha1.PostgresDB) (*v1alpha1.PostgresDB, error) {

	name := fmt.Sprintf("kubes-%s-%s", resource.Spec.Name, resource.GetUID())
	username := rds.GenerateRandomUsername(16)
	password := rds.GenerateRandomPassword(32)
	securitygroup := dbConfig.SecurityGroupId
	securitygroups := make([]*string, 0, 5)
	securitygroups = append(securitygroups, &securitygroup)
	tags := make([]*db.Tag, 0, 5)
	nskey := "Namespace"
	nsvalue := resource.Namespace
	nstag := &db.Tag{
		Key:   &nskey,
		Value: &nsvalue,
	}
	appkey := "App"
	appvalue := resource.Spec.Name
	apptag := &db.Tag{
		Key:   &appkey,
		Value: &appvalue,
	}
	tags = append(tags, nstag)
	tags = append(tags, apptag)

	input := &db.DB{
		Name:               &name,
		MasterUsername:     &username,
		MasterUserPassword: &password,
		MultiAZ:            &dbConfig.Multizone,
		SubnetGroupName:    &dbConfig.SubnetGroupName,
		SecurityGroups:     securitygroups,
		Tags:               tags,
	}
	if dbConfig.EncryptionKey != "" {
		input.KMSKeyArn = &dbConfig.EncryptionKey
	}

	// var instance *db.DB
	// var err error
	if resource.Status.ARN == "" {
		// need to create a new db
		log.Printf("Creating DB with ID: %s", name)
		log.Printf("Creating DB with master credentials: %s %s", username, password)
		// instance, err = rds.Create(input)
		// if err != nil {
		// 	return nil, err
		// }
		// resource.Status.ARN = *instance.ARN
	} else {
		// need to update the status
		// instance, err = rds.Stat(name)
		// if err != nil {
		// 	return nil, err
		// }
	}

	// resource.Status.Ready = *instance.Status
	return resource, nil
}
