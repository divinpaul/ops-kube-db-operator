package controller_test

import (
	"testing"
	"time"

	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/clientset/versioned/fake"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/client/informers/externalversions"
	"github.com/MYOB-Technology/ops-kube-db-operator/pkg/controller"
)

type mockWorker struct {
	Calls map[string][]interface{}
}

func (w *mockWorker) OnCreate(obj interface{}) {
	w.Calls["create"] = append(w.Calls["create"], obj)
}

func (w *mockWorker) OnUpdate(obj interface{}, newObj interface{}) {
	w.Calls["update"] = append(w.Calls["update"], obj)
}

func (w *mockWorker) OnDelete(obj interface{}) {
	w.Calls["delete"] = append(w.Calls["delete"], obj)
}

func newMockWorker() *mockWorker {
	return &mockWorker{
		Calls: make(map[string][]interface{}),
	}
}

func TestPgController(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	i := externalversions.NewSharedInformerFactory(clientset, time.Second*30)
	stopCh := make(chan struct{})

	i.Start(stopCh)
	wrkr := newMockWorker()

	c := controller.New(i, wrkr)
	go c.Run(stopCh)
	defer func() {
		stopCh <- struct{}{}
	}()
}
