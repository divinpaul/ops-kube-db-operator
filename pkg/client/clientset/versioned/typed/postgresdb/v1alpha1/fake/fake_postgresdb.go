/*
MYOB 2017
All Rights Reserved
*/
package fake

import (
	v1alpha1 "github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePostgresDBs implements PostgresDBInterface
type FakePostgresDBs struct {
	Fake *FakePostgresdbV1alpha1
	ns   string
}

var postgresdbsResource = schema.GroupVersionResource{Group: "postgresdb", Version: "v1alpha1", Resource: "postgresdbs"}

var postgresdbsKind = schema.GroupVersionKind{Group: "postgresdb", Version: "v1alpha1", Kind: "PostgresDB"}

// Get takes name of the postgresDB, and returns the corresponding postgresDB object, and an error if there is any.
func (c *FakePostgresDBs) Get(name string, options v1.GetOptions) (result *v1alpha1.PostgresDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(postgresdbsResource, c.ns, name), &v1alpha1.PostgresDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PostgresDB), err
}

// List takes label and field selectors, and returns the list of PostgresDBs that match those selectors.
func (c *FakePostgresDBs) List(opts v1.ListOptions) (result *v1alpha1.PostgresDBList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(postgresdbsResource, postgresdbsKind, c.ns, opts), &v1alpha1.PostgresDBList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PostgresDBList{}
	for _, item := range obj.(*v1alpha1.PostgresDBList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested postgresDBs.
func (c *FakePostgresDBs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(postgresdbsResource, c.ns, opts))

}

// Create takes the representation of a postgresDB and creates it.  Returns the server's representation of the postgresDB, and an error, if there is any.
func (c *FakePostgresDBs) Create(postgresDB *v1alpha1.PostgresDB) (result *v1alpha1.PostgresDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(postgresdbsResource, c.ns, postgresDB), &v1alpha1.PostgresDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PostgresDB), err
}

// Update takes the representation of a postgresDB and updates it. Returns the server's representation of the postgresDB, and an error, if there is any.
func (c *FakePostgresDBs) Update(postgresDB *v1alpha1.PostgresDB) (result *v1alpha1.PostgresDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(postgresdbsResource, c.ns, postgresDB), &v1alpha1.PostgresDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PostgresDB), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePostgresDBs) UpdateStatus(postgresDB *v1alpha1.PostgresDB) (*v1alpha1.PostgresDB, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(postgresdbsResource, "status", c.ns, postgresDB), &v1alpha1.PostgresDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PostgresDB), err
}

// Delete takes name of the postgresDB and deletes it. Returns an error if one occurs.
func (c *FakePostgresDBs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(postgresdbsResource, c.ns, name), &v1alpha1.PostgresDB{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePostgresDBs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(postgresdbsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.PostgresDBList{})
	return err
}

// Patch applies the patch and returns the patched postgresDB.
func (c *FakePostgresDBs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.PostgresDB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(postgresdbsResource, c.ns, name, data, subresources...), &v1alpha1.PostgresDB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PostgresDB), err
}
