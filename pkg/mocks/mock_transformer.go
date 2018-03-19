// Code generated by MockGen. DO NOT EDIT.
// Source: transformer.go

// Package mocks is a generated GoMock package.
package mocks

import (
	v1alpha1 "github.com/MYOB-Technology/ops-kube-db-operator/pkg/apis/postgresdb/v1alpha1"
	database "github.com/MYOB-Technology/ops-kube-db-operator/pkg/database"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockTransformer is a mock of Transformer interface
type MockTransformer struct {
	ctrl     *gomock.Controller
	recorder *MockTransformerMockRecorder
}

// MockTransformerMockRecorder is the mock recorder for MockTransformer
type MockTransformerMockRecorder struct {
	mock *MockTransformer
}

// NewMockTransformer creates a new mock instance
func NewMockTransformer(ctrl *gomock.Controller) *MockTransformer {
	mock := &MockTransformer{ctrl: ctrl}
	mock.recorder = &MockTransformerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTransformer) EXPECT() *MockTransformerMockRecorder {
	return m.recorder
}

// CRDToRequest mocks base method
func (m *MockTransformer) CRDToRequest(crd *v1alpha1.PostgresDB) *database.Request {
	ret := m.ctrl.Call(m, "CRDToRequest", crd)
	ret0, _ := ret[0].(*database.Request)
	return ret0
}

// CRDToRequest indicates an expected call of CRDToRequest
func (mr *MockTransformerMockRecorder) CRDToRequest(crd interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CRDToRequest", reflect.TypeOf((*MockTransformer)(nil).CRDToRequest), crd)
}
