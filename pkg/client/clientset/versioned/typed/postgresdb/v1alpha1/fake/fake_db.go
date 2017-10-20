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

// FakeDBs implements DBInterface
type FakeDBs struct {
	Fake *FakePostgresdbV1alpha1
	ns   string
}

var dbsResource = schema.GroupVersionResource{Group: "postgresdb", Version: "v1alpha1", Resource: "dbs"}

var dbsKind = schema.GroupVersionKind{Group: "postgresdb", Version: "v1alpha1", Kind: "DB"}

// Get takes name of the dB, and returns the corresponding dB object, and an error if there is any.
func (c *FakeDBs) Get(name string, options v1.GetOptions) (result *v1alpha1.DB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(dbsResource, c.ns, name), &v1alpha1.DB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DB), err
}

// List takes label and field selectors, and returns the list of DBs that match those selectors.
func (c *FakeDBs) List(opts v1.ListOptions) (result *v1alpha1.DBList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(dbsResource, dbsKind, c.ns, opts), &v1alpha1.DBList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.DBList{}
	for _, item := range obj.(*v1alpha1.DBList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested dBs.
func (c *FakeDBs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(dbsResource, c.ns, opts))

}

// Create takes the representation of a dB and creates it.  Returns the server's representation of the dB, and an error, if there is any.
func (c *FakeDBs) Create(dB *v1alpha1.DB) (result *v1alpha1.DB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(dbsResource, c.ns, dB), &v1alpha1.DB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DB), err
}

// Update takes the representation of a dB and updates it. Returns the server's representation of the dB, and an error, if there is any.
func (c *FakeDBs) Update(dB *v1alpha1.DB) (result *v1alpha1.DB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(dbsResource, c.ns, dB), &v1alpha1.DB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DB), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeDBs) UpdateStatus(dB *v1alpha1.DB) (*v1alpha1.DB, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(dbsResource, "status", c.ns, dB), &v1alpha1.DB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DB), err
}

// Delete takes name of the dB and deletes it. Returns an error if one occurs.
func (c *FakeDBs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(dbsResource, c.ns, name), &v1alpha1.DB{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDBs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(dbsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.DBList{})
	return err
}

// Patch applies the patch and returns the patched dB.
func (c *FakeDBs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.DB, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(dbsResource, c.ns, name, data, subresources...), &v1alpha1.DB{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DB), err
}
