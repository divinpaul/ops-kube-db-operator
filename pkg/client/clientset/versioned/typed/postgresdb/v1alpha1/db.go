/*
MYOB 2017
All Rights Reserved
*/
package v1alpha1

import (
	v1alpha1 "github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	scheme "github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// DBsGetter has a method to return a DBInterface.
// A group's client should implement this interface.
type DBsGetter interface {
	DBs(namespace string) DBInterface
}

// DBInterface has methods to work with DB resources.
type DBInterface interface {
	Create(*v1alpha1.DB) (*v1alpha1.DB, error)
	Update(*v1alpha1.DB) (*v1alpha1.DB, error)
	UpdateStatus(*v1alpha1.DB) (*v1alpha1.DB, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.DB, error)
	List(opts v1.ListOptions) (*v1alpha1.DBList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.DB, err error)
	DBExpansion
}

// dBs implements DBInterface
type dBs struct {
	client rest.Interface
	ns     string
}

// newDBs returns a DBs
func newDBs(c *PostgresdbV1alpha1Client, namespace string) *dBs {
	return &dBs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the dB, and returns the corresponding dB object, and an error if there is any.
func (c *dBs) Get(name string, options v1.GetOptions) (result *v1alpha1.DB, err error) {
	result = &v1alpha1.DB{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("dbs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of DBs that match those selectors.
func (c *dBs) List(opts v1.ListOptions) (result *v1alpha1.DBList, err error) {
	result = &v1alpha1.DBList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("dbs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested dBs.
func (c *dBs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("dbs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a dB and creates it.  Returns the server's representation of the dB, and an error, if there is any.
func (c *dBs) Create(dB *v1alpha1.DB) (result *v1alpha1.DB, err error) {
	result = &v1alpha1.DB{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("dbs").
		Body(dB).
		Do().
		Into(result)
	return
}

// Update takes the representation of a dB and updates it. Returns the server's representation of the dB, and an error, if there is any.
func (c *dBs) Update(dB *v1alpha1.DB) (result *v1alpha1.DB, err error) {
	result = &v1alpha1.DB{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("dbs").
		Name(dB.Name).
		Body(dB).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *dBs) UpdateStatus(dB *v1alpha1.DB) (result *v1alpha1.DB, err error) {
	result = &v1alpha1.DB{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("dbs").
		Name(dB.Name).
		SubResource("status").
		Body(dB).
		Do().
		Into(result)
	return
}

// Delete takes name of the dB and deletes it. Returns an error if one occurs.
func (c *dBs) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("dbs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *dBs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("dbs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched dB.
func (c *dBs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.DB, err error) {
	result = &v1alpha1.DB{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("dbs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
