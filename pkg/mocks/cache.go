// Code generated by MockGen. DO NOT EDIT.
// Source: cache.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// RedisClient is a mock of RedisClient interface.
type RedisClient struct {
	ctrl     *gomock.Controller
	recorder *RedisClientMockRecorder
}

// RedisClientMockRecorder is the mock recorder for RedisClient.
type RedisClientMockRecorder struct {
	mock *RedisClient
}

// NewRedisClient creates a new mock instance.
func NewRedisClient(ctrl *gomock.Controller) *RedisClient {
	mock := &RedisClient{ctrl: ctrl}
	mock.recorder = &RedisClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *RedisClient) EXPECT() *RedisClientMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *RedisClient) Delete(ctx context.Context, keys ...string) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range keys {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Delete", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *RedisClientMockRecorder) Delete(ctx interface{}, keys ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, keys...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*RedisClient)(nil).Delete), varargs...)
}

// Enabled mocks base method.
func (m *RedisClient) Enabled() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Enabled")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Enabled indicates an expected call of Enabled.
func (mr *RedisClientMockRecorder) Enabled() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enabled", reflect.TypeOf((*RedisClient)(nil).Enabled))
}

// Expire mocks base method.
func (m *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Expire", ctx, key, ttl)
	ret0, _ := ret[0].(error)
	return ret0
}

// Expire indicates an expected call of Expire.
func (mr *RedisClientMockRecorder) Expire(ctx, key, ttl interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Expire", reflect.TypeOf((*RedisClient)(nil).Expire), ctx, key, ttl)
}

// Load mocks base method.
func (m *RedisClient) Load(ctx context.Context, key string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Load", ctx, key)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Load indicates an expected call of Load.
func (mr *RedisClientMockRecorder) Load(ctx, key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Load", reflect.TypeOf((*RedisClient)(nil).Load), ctx, key)
}

// LoadMany mocks base method.
func (m *RedisClient) LoadMany(ctx context.Context, keys ...string) ([]string, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range keys {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "LoadMany", varargs...)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LoadMany indicates an expected call of LoadMany.
func (mr *RedisClientMockRecorder) LoadMany(ctx interface{}, keys ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, keys...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadMany", reflect.TypeOf((*RedisClient)(nil).LoadMany), varargs...)
}

// Store mocks base method.
func (m *RedisClient) Store(ctx context.Context, key string, value any, ttl time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", ctx, key, value, ttl)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store.
func (mr *RedisClientMockRecorder) Store(ctx, key, value, ttl interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*RedisClient)(nil).Store), ctx, key, value, ttl)
}
