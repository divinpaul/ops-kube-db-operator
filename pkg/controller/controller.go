package controller

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"

	"github.com/kr/pretty"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/workqueue"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// dbConfig defines default DB instance specifics
	dbConfig *db.DB
}

// New instantiates an rds controller
func New(
	client *kubernetes.Clientset,
	dbClient dbClientSet.Interface,
	dbInformer dbInformerFactory.SharedInformerFactory,
) (*RDSController, error) {

	informer := dbInformer.Postgresdb().V1alpha1().PostgresDBs()
	dbConfig, err := NewDBInstanceConfig(client)
	if err != nil {
		return nil, fmt.Errorf("error reading configmap: %s", err)
	}

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

	return c, nil
}

// NewDBInstanceConfig defines the required minimum settings for the instance
func NewDBInstanceConfig(client *kubernetes.Clientset) (*db.DB, error) {

	cfgmaps := client.ConfigMaps("kube-system")
	cfg, err := cfgmaps.Get("ops-kube-db-operator", meta_v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error reading configmap: %s", err)
	}

	// required config
	dbSubnetGroupName := cfg.Data["aws.subnet.group.name"]
	dbSecurityGroupID := cfg.Data["aws.security.group.id"]

	if dbSecurityGroupID == "" {
		return nil, fmt.Errorf("error: required environment variable DB_SECURITYGROUPID missing")
	}
	if dbSubnetGroupName == "" {
		return nil, fmt.Errorf("error: required environment variable DB_SUBNETGROUPNAME missing")
	}

	securitygroups := make([]*string, 0, 5)
	securitygroups = append(securitygroups, &dbSecurityGroupID)
	dbConfig := db.DB{
		SubnetGroupName: &dbSubnetGroupName,
		SecurityGroups:  securitygroups,
	}

	// retrieve db port if provided
	var dbPort int64
	if cfg.Data["aws.rds.default.postgres.default.port"] == "" {
		dbPort = 5432
	} else {
		dbPort, err = strconv.ParseInt(cfg.Data["aws.rds.postgres.default.port"], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing port : %s", cfg.Data["aws.rds.postgres.default.port"])
		}
	}
	dbConfig.Port = &dbPort

	// retrieve multiaz bool if provided
	dbMultiAz := true
	if cfg.Data["aws.rds.postgres.default.multiaz"] != "" {
		dbMultiAz, err = strconv.ParseBool(cfg.Data["aws.rds.postgres.default.multiaz"])
		if err != nil {
			fmt.Printf("warning: error parsing multiaz , default to false: %s", cfg.Data["aws.rds.postgres.default.multiaz"])
		}
	}
	dbConfig.MultiAZ = &dbMultiAz

	// retrieve kms encryption bool if provided
	dbKmsEncryption := true
	if cfg.Data["aws.rds.postgres.default.kms.encryption"] != "" {
		dbKmsEncryption, err = strconv.ParseBool(cfg.Data["aws.rds.postgres.default.kms.encryption"])
		if err != nil {
			fmt.Printf("warning: error parsing kms encryption , default to true: %s", cfg.Data["aws.rds.postgres.default.kms.encryption"])
			dbConfig.StorageEncrypted = &dbKmsEncryption
		}
	}

	// retrieve kms encryption key arn if provided
	var dbKmsKeyArn string
	if cfg.Data["aws.kms.key.arn"] != "" {
		dbKmsKeyArn = cfg.Data["aws.kms.key.arn"]
		dbConfig.KMSKeyArn = &dbKmsKeyArn
	}

	// retrieve default instance class if provided
	var dbInstanceClass string
	if cfg.Data["aws.rds.postgres.default.instance.class"] != "" {
		dbInstanceClass = cfg.Data["aws.rds.postgres.default.instance.class"]
		dbConfig.DBInstanceClass = &dbInstanceClass
	}

	// retrieve preferred backup window if provided
	var dbBackupWindow string
	if cfg.Data["aws.rds.postgres.default.backup.window"] != "" {
		dbBackupWindow = cfg.Data["aws.rds.postgres.default.backup.window"]
		dbConfig.PreferredBackupWindow = &dbBackupWindow
	}

	// retrieve preferred maintenance window if provided
	var dbMaintenanceWindow string
	if cfg.Data["aws.rds.postgres.default.maintenance.window"] != "" {
		dbMaintenanceWindow = cfg.Data["aws.rds.postgres.default.maintenance.window"]
		dbConfig.PreferredMaintenanceWindow = &dbMaintenanceWindow
	}

	// retrieve default storage class
	var dbStorageType string
	if cfg.Data["aws.rds.postgres.default.storage.type"] != "" {
		dbStorageType = cfg.Data["aws.rds.postgres.default.storage.type"]
		dbConfig.StorageType = &dbStorageType
	}

	// retrieve default storage size
	var dbStorageSize int64
	if cfg.Data["aws.rds.postgres.default.storage.size"] != "" {
		dbStorageSize, err = strconv.ParseInt(cfg.Data["aws.rds.postgres.default.storage.size"], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing storage size : %s", cfg.Data["aws.rds.postgres.default.storage.size:"])
		}
		dbConfig.StorageAllocatedGB = &dbStorageSize
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

	fmt.Printf("work item: %# v", pretty.Formatter(key))

	// indicate to queue when work is finished on a specific item
	defer c.queue.Done(key)

	err := c.processDB(key.(string))
	if err == nil {
		// processed successfully, lets forget item in queue and return success
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
	fmt.Printf("db: %# v", pretty.Formatter(db))

	// deep copy to not change the cache
	newDbInterface, _ := scheme.Scheme.DeepCopy(db)
	newDb := newDbInterface.(*v1alpha1.PostgresDB)
	instanceID := fmt.Sprintf("%s-%s", name, newDb.GetUID())
	// ensure instanceId is less than 63 chars
	instanceID = fmt.Sprintf("%.63s", instanceID)
	c.dbConfig.Name = &instanceID

	// list the db and get status
	instance, err := c.rds.Stat(instanceID)
	if err != nil {
		fmt.Printf("%s: %s\n", name, err)
		//  TODO test for instance not found error
		//		return fmt.Errorf("failed querying db %s: requeuing - %v", instanceId, err)
	}

	fmt.Printf("instance: %# v", pretty.Formatter(instance))

	// otherwise create the db now
	newObj, err := c.createDb(c.dbConfig, newDb)
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

func (c *RDSController) configureDB(dbConfig *db.DB, resource *v1alpha1.PostgresDB) *db.DB {

	// TODO - store in a secret
	username := c.rds.GenerateRandomUsername(16)
	password := c.rds.GenerateRandomPassword(32)

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

	dbConfig.MasterUsername = &username
	dbConfig.MasterUserPassword = &password
	dbConfig.Tags = tags
	if resource.Spec.Size != "" {
		dbConfig.DBInstanceClass = &resource.Spec.Size
	}

	// retrieve gigabytes if provided
	var dbStorageAllocatedGB int64
	if resource.Spec.GigaBytes > 0 {
		dbStorageAllocatedGB = resource.Spec.GigaBytes
		dbConfig.StorageAllocatedGB = &dbStorageAllocatedGB
	}

	// retrieve iops if provided
	var dbStorageIops int64
	var dbStorageType string
	if resource.Spec.Iops > 0 {
		dbStorageIops = resource.Spec.Iops
		dbConfig.StorageIops = &dbStorageIops
		// if iops is provided, explicitly set the storage type to io1
		dbStorageType = "io1"
		dbConfig.StorageType = &dbStorageType
	}

	return dbConfig

}

func (c *RDSController) createDb(dbConfig *db.DB, resource *v1alpha1.PostgresDB) (*v1alpha1.PostgresDB, error) {

	if resource.Status.ARN == "" {

		// need to create a new db
		input := c.configureDB(dbConfig, resource)
		log.Printf("Creating DB with ID: %s", *input.Name)
		log.Printf("Creating DB with master credentials: %s %s", *input.MasterUsername, *input.MasterUserPassword)
		instance, err := c.rds.Create(input)
		if err != nil {
			return nil, err
		}
		// need to update the status
		instance, err = c.rds.Stat(*dbConfig.Name)
		if err != nil {
			return nil, err
		}
		resource.Status.ARN = *instance.ARN
	}

	// resource.Status.Ready = *instance.Status
	return resource, nil
}
