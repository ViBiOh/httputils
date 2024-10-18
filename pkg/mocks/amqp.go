// Code generated by MockGen. DO NOT EDIT.
// Source: amqp.go
//
// Generated by this command:
//
//	mockgen -source amqp.go -destination ../mocks/amqp.go -package mocks -mock_names Connection=AMQPConnection
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	amqp091 "github.com/rabbitmq/amqp091-go"
	gomock "go.uber.org/mock/gomock"
)

// AMQPConnection is a mock of Connection interface.
type AMQPConnection struct {
	isgomock struct{}
	ctrl     *gomock.Controller
	recorder *AMQPConnectionMockRecorder
}

// AMQPConnectionMockRecorder is the mock recorder for AMQPConnection.
type AMQPConnectionMockRecorder struct {
	mock *AMQPConnection
}

// NewAMQPConnection creates a new mock instance.
func NewAMQPConnection(ctrl *gomock.Controller) *AMQPConnection {
	mock := &AMQPConnection{ctrl: ctrl}
	mock.recorder = &AMQPConnectionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *AMQPConnection) EXPECT() *AMQPConnectionMockRecorder {
	return m.recorder
}

// Channel mocks base method.
func (m *AMQPConnection) Channel() (*amqp091.Channel, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Channel")
	ret0, _ := ret[0].(*amqp091.Channel)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Channel indicates an expected call of Channel.
func (mr *AMQPConnectionMockRecorder) Channel() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Channel", reflect.TypeOf((*AMQPConnection)(nil).Channel))
}

// Close mocks base method.
func (m *AMQPConnection) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *AMQPConnectionMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*AMQPConnection)(nil).Close))
}

// IsClosed mocks base method.
func (m *AMQPConnection) IsClosed() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsClosed")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsClosed indicates an expected call of IsClosed.
func (mr *AMQPConnectionMockRecorder) IsClosed() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsClosed", reflect.TypeOf((*AMQPConnection)(nil).IsClosed))
}
