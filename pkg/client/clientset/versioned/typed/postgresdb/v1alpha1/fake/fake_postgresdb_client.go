/*
MYOB 2017
All Rights Reserved
*/
package fake

import (
	v1alpha1 "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned/typed/postgresdb/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakePostgresdbV1alpha1 struct {
	*testing.Fake
}

func (c *FakePostgresdbV1alpha1) PostgresDBs(namespace string) v1alpha1.PostgresDBInterface {
	return &FakePostgresDBs{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakePostgresdbV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
