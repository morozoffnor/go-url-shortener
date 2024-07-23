// Code generated by MockGen. DO NOT EDIT.
// Source: storage.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	storage "github.com/morozoffnor/go-url-shortener/internal/storage"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// AddBatch mocks base method.
func (m *MockStorage) AddBatch(ctx context.Context, urls []storage.BatchInput) ([]storage.BatchOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddBatch", ctx, urls)
	ret0, _ := ret[0].([]storage.BatchOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddBatch indicates an expected call of AddBatch.
func (mr *MockStorageMockRecorder) AddBatch(ctx, urls interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddBatch", reflect.TypeOf((*MockStorage)(nil).AddBatch), ctx, urls)
}

// AddNewURL mocks base method.
func (m *MockStorage) AddNewURL(ctx context.Context, full string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddNewURL", ctx, full)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddNewURL indicates an expected call of AddNewURL.
func (mr *MockStorageMockRecorder) AddNewURL(ctx, full interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddNewURL", reflect.TypeOf((*MockStorage)(nil).AddNewURL), ctx, full)
}

// GetFullURL mocks base method.
func (m *MockStorage) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFullURL", ctx, shortURL)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFullURL indicates an expected call of GetFullURL.
func (mr *MockStorageMockRecorder) GetFullURL(ctx, shortURL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFullURL", reflect.TypeOf((*MockStorage)(nil).GetFullURL), ctx, shortURL)
}

// Ping mocks base method.
func (m *MockStorage) Ping(ctx context.Context) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping", ctx)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockStorageMockRecorder) Ping(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockStorage)(nil).Ping), ctx)
}
