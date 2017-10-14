/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package fake

import (
	v1alpha1 "github.com/gugahoi/rds-operator/pkg/apis/db/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeDBs implements DBInterface
type FakeDBs struct {
	Fake *FakeDbV1alpha1
	ns   string
}

var dbsResource = schema.GroupVersionResource{Group: "db", Version: "v1alpha1", Resource: "dbs"}

var dbsKind = schema.GroupVersionKind{Group: "db", Version: "v1alpha1", Kind: "DB"}

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
