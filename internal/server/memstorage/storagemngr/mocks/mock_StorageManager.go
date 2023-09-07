// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/erupshis/metrics/internal/server/memstorage/storagemngr (interfaces: StorageManager)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockStorageManager is a mock of StorageManager interface.
type MockStorageManager struct {
	ctrl     *gomock.Controller
	recorder *MockStorageManagerMockRecorder
}

// MockStorageManagerMockRecorder is the mock recorder for MockStorageManager.
type MockStorageManagerMockRecorder struct {
	mock *MockStorageManager
}

// NewMockStorageManager creates a new mock instance.
func NewMockStorageManager(ctrl *gomock.Controller) *MockStorageManager {
	mock := &MockStorageManager{ctrl: ctrl}
	mock.recorder = &MockStorageManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorageManager) EXPECT() *MockStorageManagerMockRecorder {
	return m.recorder
}

// CheckConnection mocks base method.
func (m *MockStorageManager) CheckConnection(arg0 context.Context) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckConnection", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckConnection indicates an expected call of CheckConnection.
func (mr *MockStorageManagerMockRecorder) CheckConnection(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckConnection", reflect.TypeOf((*MockStorageManager)(nil).CheckConnection), arg0)
}

// Close mocks base method.
func (m *MockStorageManager) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockStorageManagerMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStorageManager)(nil).Close))
}

// RestoreDataFromStorage mocks base method.
func (m *MockStorageManager) RestoreDataFromStorage(arg0 context.Context) (map[string]float64, map[string]int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RestoreDataFromStorage", arg0)
	ret0, _ := ret[0].(map[string]float64)
	ret1, _ := ret[1].(map[string]int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// RestoreDataFromStorage indicates an expected call of RestoreDataFromStorage.
func (mr *MockStorageManagerMockRecorder) RestoreDataFromStorage(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RestoreDataFromStorage", reflect.TypeOf((*MockStorageManager)(nil).RestoreDataFromStorage), arg0)
}

// SaveMetricsInStorage mocks base method.
func (m *MockStorageManager) SaveMetricsInStorage(arg0 context.Context, arg1, arg2 map[string]interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveMetricsInStorage", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveMetricsInStorage indicates an expected call of SaveMetricsInStorage.
func (mr *MockStorageManagerMockRecorder) SaveMetricsInStorage(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveMetricsInStorage", reflect.TypeOf((*MockStorageManager)(nil).SaveMetricsInStorage), arg0, arg1, arg2)
}
