package pgdb

import (
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"

	dfm "github.com/MYOB-Technology/dataform/pkg/db"
	dfmsvc "github.com/MYOB-Technology/dataform/pkg/service"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	dbClientSet "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned"
	dbLister "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/listers/postgresdb/v1alpha1"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
)

// Manager sets up defaults and config for new PgDB objects
type Manager struct {
	// client is the standart kubernetes clientset
	client *kubernetes.Clientset
	// dbClient is the client for the DB crd
	dbClient dbClientSet.Interface
	// dbLister is the cache of DBs used for lookup
	lister dbLister.PostgresDBLister
	// defaults holds the rds db default settings
	defaults *dfm.DB
	rds      *dfm.Manager
}

// NewManager creates a new Manager object
func NewManager(client *kubernetes.Clientset, dbClient dbClientSet.Interface, lister dbLister.PostgresDBLister) (*Manager, error) {
	defaults, err := NewDBDefaults(client)
	if err != nil {
		return nil, err
	}
	return &Manager{
		client:   client,
		dbClient: dbClient,
		lister:   lister,
		defaults: defaults,
		rds:      dfm.NewManager(dfmsvc.New("")),
	}, nil
}

// Sync will test for existence and create and save postgresdb resources
func (p *Manager) Sync(key string) error {
	// split the namespace and name from cache

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("error splitting namespace/key from obj %s: %v", key, err)
	}
	log.Infof("sync processing item: %s/%s", ns, name)

	resource, err := p.lister.PostgresDBs(ns).Get(name)
	if err != nil {
		return fmt.Errorf("error: sync failed to retrieve up to date db resource %s, it has most likely been deleted: %v", key, err)
	}

	// deep copy to not change the cache
	newDbInterface, _ := scheme.Scheme.DeepCopy(resource)
	newDb := newDbInterface.(*v1alpha1.PostgresDB)

	instanceID := fmt.Sprintf("%s-%s", name, newDb.GetUID())
	// ensure instanceId is less than 63 chars
	instanceID = fmt.Sprintf("%.63s", instanceID)

	instance := &dfm.DB{}
	*instance = *p.defaults
	instance.Name = &instanceID
	pgdb := &PgDB{
		exists:   true,
		ns:       ns,
		klient:   p.client,
		dbklient: p.dbClient,
		obj:      resource,
		db:       instance,
		rds:      p.rds,
	}

	return pgdb.Save()
}

// NewDBDefaults builds the default DB object
func NewDBDefaults(client *kubernetes.Clientset) (*dfm.DB, error) {

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
	dbConfig := dfm.DB{
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
