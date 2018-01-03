package controller

import (
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/informers/externalversions"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/listers/postgresdb/v1alpha1"
	"github.com/golang/glog"
	"k8s.io/client-go/tools/cache"
)

type Worker interface {
	OnCreate(obj interface{})
	OnUpdate(obj interface{}, newObj interface{})
	OnDelete(obj interface{})
}

// PgController is a controller for Postgres RDS DBs.
type PgController struct {
	dbsLister v1alpha1.PostgresDBLister
	dbsSynced cache.InformerSynced
}

// New instantiates an pgController
func New(factory externalversions.SharedInformerFactory, worker Worker) *PgController {

	informer := factory.Postgresdb().V1alpha1().PostgresDBs()
	c := &PgController{
		dbsLister: informer.Lister(),
		dbsSynced: informer.Informer().HasSynced,
	}

	// Just Call worker function rather then add, update, delete
	informer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    worker.OnCreate,
			UpdateFunc: worker.OnUpdate,
			DeleteFunc: worker.OnDelete,
		},
	)
	return c
}

func (c *PgController) Run(stopCh <-chan struct{}) {
	glog.Info("starting the controller")
	if !cache.WaitForCacheSync(stopCh, c.dbsSynced) {
		glog.Info("unable to sync cache")
		return
	}
	glog.Info("caches are synced")

	// wait until we're told to stop
	glog.Info("waiting for stop signal")
	<-stopCh
	glog.Info("received stop signal")
}
