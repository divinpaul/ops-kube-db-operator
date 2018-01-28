/*

Copyright 2017 MYOB Technology Pty Ltd

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
documentation files (the "Software"), to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

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

// PostgresDBsGetter has a method to return a PostgresDBInterface.
// A group's client should implement this interface.
type PostgresDBsGetter interface {
	PostgresDBs(namespace string) PostgresDBInterface
}

// PostgresDBInterface has methods to work with PostgresDB resources.
type PostgresDBInterface interface {
	Create(*v1alpha1.PostgresDB) (*v1alpha1.PostgresDB, error)
	Update(*v1alpha1.PostgresDB) (*v1alpha1.PostgresDB, error)
	UpdateStatus(*v1alpha1.PostgresDB) (*v1alpha1.PostgresDB, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.PostgresDB, error)
	List(opts v1.ListOptions) (*v1alpha1.PostgresDBList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.PostgresDB, err error)
	PostgresDBExpansion
}

// postgresDBs implements PostgresDBInterface
type postgresDBs struct {
	client rest.Interface
	ns     string
}

// newPostgresDBs returns a PostgresDBs
func newPostgresDBs(c *PostgresdbV1alpha1Client, namespace string) *postgresDBs {
	return &postgresDBs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the postgresDB, and returns the corresponding postgresDB object, and an error if there is any.
func (c *postgresDBs) Get(name string, options v1.GetOptions) (result *v1alpha1.PostgresDB, err error) {
	result = &v1alpha1.PostgresDB{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("postgresdbs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of PostgresDBs that match those selectors.
func (c *postgresDBs) List(opts v1.ListOptions) (result *v1alpha1.PostgresDBList, err error) {
	result = &v1alpha1.PostgresDBList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("postgresdbs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested postgresDBs.
func (c *postgresDBs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("postgresdbs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a postgresDB and creates it.  Returns the server's representation of the postgresDB, and an error, if there is any.
func (c *postgresDBs) Create(postgresDB *v1alpha1.PostgresDB) (result *v1alpha1.PostgresDB, err error) {
	result = &v1alpha1.PostgresDB{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("postgresdbs").
		Body(postgresDB).
		Do().
		Into(result)
	return
}

// Update takes the representation of a postgresDB and updates it. Returns the server's representation of the postgresDB, and an error, if there is any.
func (c *postgresDBs) Update(postgresDB *v1alpha1.PostgresDB) (result *v1alpha1.PostgresDB, err error) {
	result = &v1alpha1.PostgresDB{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("postgresdbs").
		Name(postgresDB.Name).
		Body(postgresDB).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *postgresDBs) UpdateStatus(postgresDB *v1alpha1.PostgresDB) (result *v1alpha1.PostgresDB, err error) {
	result = &v1alpha1.PostgresDB{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("postgresdbs").
		Name(postgresDB.Name).
		SubResource("status").
		Body(postgresDB).
		Do().
		Into(result)
	return
}

// Delete takes name of the postgresDB and deletes it. Returns an error if one occurs.
func (c *postgresDBs) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("postgresdbs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *postgresDBs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("postgresdbs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched postgresDB.
func (c *postgresDBs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.PostgresDB, err error) {
	result = &v1alpha1.PostgresDB{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("postgresdbs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
