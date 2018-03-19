package k8s

import (
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/database"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CRDClient struct {
	client versioned.Interface
}

func NewCRDClient(c versioned.Interface) *CRDClient {
	return &CRDClient{
		client: c,
	}
}

func (u *CRDClient) StatusUpdate(sReq *database.StatusRequest) error {
	crd, err := u.client.PostgresdbV1alpha1().PostgresDBs(string(sReq.Scope)).Get(sReq.Name, v1.GetOptions{})
	if err != nil {
		return err
	}

	status := &v1alpha1.PostgresDBStatus{Ready: database.GetMessageForStatus(sReq.Status)}
	if sReq.ID != nil {
		status.ID = string(*sReq.ID)
	}

	crd.Status = *status

	_, err = u.client.PostgresdbV1alpha1().PostgresDBs(string(sReq.Scope)).Update(crd)
	if err != nil {
		return err
	}
	return nil
}
