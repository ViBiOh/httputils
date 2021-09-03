// Code generated by MockGen. DO NOT EDIT.
// Source: io (interfaces: ReadCloser)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// ReadCloser is a mock of ReadCloser interface.
type ReadCloser struct {
	ctrl     *gomock.Controller
	recorder *ReadCloserMockRecorder
}

// ReadCloserMockRecorder is the mock recorder for ReadCloser.
type ReadCloserMockRecorder struct {
	mock *ReadCloser
}

// NewReadCloser creates a new mock instance.
func NewReadCloser(ctrl *gomock.Controller) *ReadCloser {
	mock := &ReadCloser{ctrl: ctrl}
	mock.recorder = &ReadCloserMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *ReadCloser) EXPECT() *ReadCloserMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *ReadCloser) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *ReadCloserMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*ReadCloser)(nil).Close))
}

// Read mocks base method.
func (m *ReadCloser) Read(arg0 []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *ReadCloserMockRecorder) Read(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*ReadCloser)(nil).Read), arg0)
}