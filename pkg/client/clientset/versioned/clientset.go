/*
MYOB 2017
All Rights Reserved
*/
package versioned

import (
	glog "github.com/golang/glog"
	postgresdbv1alpha1 "github.com/myob-technology/ops-kube-db-operator/pkg/client/clientset/versioned/typed/postgresdb/v1alpha1"
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	PostgresdbV1alpha1() postgresdbv1alpha1.PostgresdbV1alpha1Interface
	// Deprecated: please explicitly pick a version if possible.
	Postgresdb() postgresdbv1alpha1.PostgresdbV1alpha1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	postgresdbV1alpha1 *postgresdbv1alpha1.PostgresdbV1alpha1Client
}

// PostgresdbV1alpha1 retrieves the PostgresdbV1alpha1Client
func (c *Clientset) PostgresdbV1alpha1() postgresdbv1alpha1.PostgresdbV1alpha1Interface {
	return c.postgresdbV1alpha1
}

// Deprecated: Postgresdb retrieves the default version of PostgresdbClient.
// Please explicitly pick a version.
func (c *Clientset) Postgresdb() postgresdbv1alpha1.PostgresdbV1alpha1Interface {
	return c.postgresdbV1alpha1
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.postgresdbV1alpha1, err = postgresdbv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		glog.Errorf("failed to create the DiscoveryClient: %v", err)
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.postgresdbV1alpha1 = postgresdbv1alpha1.NewForConfigOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.postgresdbV1alpha1 = postgresdbv1alpha1.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
