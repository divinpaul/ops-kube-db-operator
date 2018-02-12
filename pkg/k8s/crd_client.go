package k8s

import (
	crds "github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	postgresdbv1alpha1 "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned/typed/postgresdb/v1alpha1"
)

type CRDClient interface {
	Update(*crds.PostgresDB) (*crds.PostgresDB, error)
}

type K8SCRDClient struct {
	clientSet postgresdbv1alpha1.PostgresdbV1alpha1Interface
}

func (this *K8SCRDClient) Update(crd *crds.PostgresDB) (*crds.PostgresDB, error) {
	return this.clientSet.PostgresDBs(crd.ObjectMeta.Namespace).Update(crd)
}

func NewK8SCRDClient(clientSet postgresdbv1alpha1.PostgresdbV1alpha1Interface) CRDClient {
	return &K8SCRDClient{
		clientSet: clientSet,
	}
}
