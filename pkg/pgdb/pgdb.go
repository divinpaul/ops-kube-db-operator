package pgdb

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	dfm "github.com/MYOB-Technology/dataform/pkg/db"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	dbClientSet "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/secret"
	"k8s.io/client-go/kubernetes"
)

// PgDB represents a Kubernetes PostgresDB resource
type PgDB struct {
	obj      *v1alpha1.PostgresDB
	klient   *kubernetes.Clientset
	dbklient dbClientSet.Interface
	exists   bool
	ns       string
	db       *dfm.DB
	rds      *dfm.Manager
}

// Save updates a postgresdb when it exists, creates a new one if it doesnt
func (p *PgDB) Save() error {
	var err error

	if p.obj.Status.ARN == "" {
		// TODO rds update
		p.configureNewDB()
		log.Infof("creating DB with ID: %s", *p.db.Name)
		log.Infof("creating DB with master credentials: %s %s", *p.db.MasterUsername, *p.db.MasterUserPassword)
		p.db, err = p.rds.Create(p.db)
		if err != nil {
			return err
		}
	}
	return p.Stat()
}

// Stat checks for existence of RDS instance and updates db info and kubes resource Status
func (p *PgDB) Stat() error {
	// list the db and get status
	instance, err := p.rds.Stat(*p.db.Name)
	if err != nil {
		log.Infof("stat instance not found %s: %v", *p.db.Name, err)
		return err
	}
	log.Infof("instance found %s: %s", *p.db.Name, *instance.ARN)

	p.obj.Status.Ready = *instance.Status
	p.obj.Status.ARN = *instance.ARN
	p.db = instance
	log.Infof("stat updating postgresdb resource %s/%s", p.ns, p.obj.ObjectMeta.Name)
	var obj *v1alpha1.PostgresDB
	obj, err = p.dbklient.Postgresdb().PostgresDBs(p.ns).Update(p.obj)
	if err != nil {
		log.Errorf("error stat updating postgresdb resource %s/%s: %v", p.ns, p.obj.ObjectMeta.Name, err)
		return err
	}
	p.obj = obj
	if p.obj.Status.Ready != "available" {
		return fmt.Errorf("postgresdb %s/%s is not yet available: %s", p.ns, p.obj.ObjectMeta.Name, p.obj.Status.Ready)
	}
	log.Infof("saved postgresdb %s/%s, status: %s", p.ns, p.obj.ObjectMeta.Name, p.obj.Status.Ready)
	return nil
}

// Delete deletes a postgresdb resource from Kubernetes
func (p *PgDB) Delete() error {
	log.Infof("delete postgresdb %s/%s", p.ns, p.obj.ObjectMeta.Name)
	// p.obj is expected to not exist - just check for  rds existence
	if err := p.Stat(); err != nil {
		_, err := p.rds.Delete(*p.db.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PgDB) configureNewDB() {

	username := p.rds.GenerateRandomUsername(16)
	password := p.rds.GenerateRandomPassword(32)

	// create secret with some info
	defer func() {
		newSec := secret.New(p.klient, p.obj.Namespace, p.obj.Name).SetData(username, password)
		newSec.Save()
	}()

	tags := make([]*dfm.Tag, 0, 5)
	nskey := "Namespace"
	nsvalue := p.obj.Namespace
	nstag := &dfm.Tag{
		Key:   &nskey,
		Value: &nsvalue,
	}
	tags = append(tags, nstag)

	p.db.MasterUsername = &username
	p.db.MasterUserPassword = &password
	p.db.Tags = tags
	if p.obj.Spec.Size != "" {
		p.db.DBInstanceClass = &p.obj.Spec.Size
	}

	// retrieve gigabytes if provided
	var dbStorageAllocatedGB int64
	if p.obj.Spec.Storage > 0 {
		dbStorageAllocatedGB = p.obj.Spec.Storage
		p.db.StorageAllocatedGB = &dbStorageAllocatedGB
	}

	// retrieve iops if provided
	var dbStorageIops int64
	var dbStorageType string
	if p.obj.Spec.Iops > 0 {
		dbStorageIops = p.obj.Spec.Iops
		p.db.StorageIops = &dbStorageIops
		// if iops is provided, explicitly set the storage type to io1
		dbStorageType = "io1"
		p.db.StorageType = &dbStorageType
	}

}
