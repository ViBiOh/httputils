// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/jackc/pgx/v5 (interfaces: Tx,Row,Rows)
//
// Generated by this command:
//
//	mockgen -destination pkg/mocks/pgx.go -package mocks -mock_names Tx=Tx,Row=Row,Rows=Rows github.com/jackc/pgx/v5 Tx,Row,Rows
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	pgx "github.com/jackc/pgx/v5"
	pgconn "github.com/jackc/pgx/v5/pgconn"
	gomock "go.uber.org/mock/gomock"
)

// Tx is a mock of Tx interface.
type Tx struct {
	isgomock struct{}
	ctrl     *gomock.Controller
	recorder *TxMockRecorder
}

// TxMockRecorder is the mock recorder for Tx.
type TxMockRecorder struct {
	mock *Tx
}

// NewTx creates a new mock instance.
func NewTx(ctrl *gomock.Controller) *Tx {
	mock := &Tx{ctrl: ctrl}
	mock.recorder = &TxMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Tx) EXPECT() *TxMockRecorder {
	return m.recorder
}

// Begin mocks base method.
func (m *Tx) Begin(ctx context.Context) (pgx.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Begin", ctx)
	ret0, _ := ret[0].(pgx.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Begin indicates an expected call of Begin.
func (mr *TxMockRecorder) Begin(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Begin", reflect.TypeOf((*Tx)(nil).Begin), ctx)
}

// Commit mocks base method.
func (m *Tx) Commit(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Commit indicates an expected call of Commit.
func (mr *TxMockRecorder) Commit(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*Tx)(nil).Commit), ctx)
}

// Conn mocks base method.
func (m *Tx) Conn() *pgx.Conn {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Conn")
	ret0, _ := ret[0].(*pgx.Conn)
	return ret0
}

// Conn indicates an expected call of Conn.
func (mr *TxMockRecorder) Conn() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Conn", reflect.TypeOf((*Tx)(nil).Conn))
}

// CopyFrom mocks base method.
func (m *Tx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CopyFrom", ctx, tableName, columnNames, rowSrc)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CopyFrom indicates an expected call of CopyFrom.
func (mr *TxMockRecorder) CopyFrom(ctx, tableName, columnNames, rowSrc any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CopyFrom", reflect.TypeOf((*Tx)(nil).CopyFrom), ctx, tableName, columnNames, rowSrc)
}

// Exec mocks base method.
func (m *Tx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, sql}
	for _, a := range arguments {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(pgconn.CommandTag)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec.
func (mr *TxMockRecorder) Exec(ctx, sql any, arguments ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, sql}, arguments...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*Tx)(nil).Exec), varargs...)
}

// LargeObjects mocks base method.
func (m *Tx) LargeObjects() pgx.LargeObjects {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LargeObjects")
	ret0, _ := ret[0].(pgx.LargeObjects)
	return ret0
}

// LargeObjects indicates an expected call of LargeObjects.
func (mr *TxMockRecorder) LargeObjects() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LargeObjects", reflect.TypeOf((*Tx)(nil).LargeObjects))
}

// Prepare mocks base method.
func (m *Tx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Prepare", ctx, name, sql)
	ret0, _ := ret[0].(*pgconn.StatementDescription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Prepare indicates an expected call of Prepare.
func (mr *TxMockRecorder) Prepare(ctx, name, sql any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Prepare", reflect.TypeOf((*Tx)(nil).Prepare), ctx, name, sql)
}

// Query mocks base method.
func (m *Tx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, sql}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(pgx.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query.
func (mr *TxMockRecorder) Query(ctx, sql any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, sql}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*Tx)(nil).Query), varargs...)
}

// QueryRow mocks base method.
func (m *Tx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	m.ctrl.T.Helper()
	varargs := []any{ctx, sql}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "QueryRow", varargs...)
	ret0, _ := ret[0].(pgx.Row)
	return ret0
}

// QueryRow indicates an expected call of QueryRow.
func (mr *TxMockRecorder) QueryRow(ctx, sql any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, sql}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRow", reflect.TypeOf((*Tx)(nil).QueryRow), varargs...)
}

// Rollback mocks base method.
func (m *Tx) Rollback(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Rollback", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Rollback indicates an expected call of Rollback.
func (mr *TxMockRecorder) Rollback(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rollback", reflect.TypeOf((*Tx)(nil).Rollback), ctx)
}

// SendBatch mocks base method.
func (m *Tx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendBatch", ctx, b)
	ret0, _ := ret[0].(pgx.BatchResults)
	return ret0
}

// SendBatch indicates an expected call of SendBatch.
func (mr *TxMockRecorder) SendBatch(ctx, b any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendBatch", reflect.TypeOf((*Tx)(nil).SendBatch), ctx, b)
}

// Row is a mock of Row interface.
type Row struct {
	isgomock struct{}
	ctrl     *gomock.Controller
	recorder *RowMockRecorder
}

// RowMockRecorder is the mock recorder for Row.
type RowMockRecorder struct {
	mock *Row
}

// NewRow creates a new mock instance.
func NewRow(ctrl *gomock.Controller) *Row {
	mock := &Row{ctrl: ctrl}
	mock.recorder = &RowMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Row) EXPECT() *RowMockRecorder {
	return m.recorder
}

// Scan mocks base method.
func (m *Row) Scan(dest ...any) error {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range dest {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Scan", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Scan indicates an expected call of Scan.
func (mr *RowMockRecorder) Scan(dest ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Scan", reflect.TypeOf((*Row)(nil).Scan), dest...)
}

// Rows is a mock of Rows interface.
type Rows struct {
	isgomock struct{}
	ctrl     *gomock.Controller
	recorder *RowsMockRecorder
}

// RowsMockRecorder is the mock recorder for Rows.
type RowsMockRecorder struct {
	mock *Rows
}

// NewRows creates a new mock instance.
func NewRows(ctrl *gomock.Controller) *Rows {
	mock := &Rows{ctrl: ctrl}
	mock.recorder = &RowsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Rows) EXPECT() *RowsMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *Rows) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *RowsMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*Rows)(nil).Close))
}

// CommandTag mocks base method.
func (m *Rows) CommandTag() pgconn.CommandTag {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CommandTag")
	ret0, _ := ret[0].(pgconn.CommandTag)
	return ret0
}

// CommandTag indicates an expected call of CommandTag.
func (mr *RowsMockRecorder) CommandTag() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CommandTag", reflect.TypeOf((*Rows)(nil).CommandTag))
}

// Conn mocks base method.
func (m *Rows) Conn() *pgx.Conn {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Conn")
	ret0, _ := ret[0].(*pgx.Conn)
	return ret0
}

// Conn indicates an expected call of Conn.
func (mr *RowsMockRecorder) Conn() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Conn", reflect.TypeOf((*Rows)(nil).Conn))
}

// Err mocks base method.
func (m *Rows) Err() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Err")
	ret0, _ := ret[0].(error)
	return ret0
}

// Err indicates an expected call of Err.
func (mr *RowsMockRecorder) Err() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Err", reflect.TypeOf((*Rows)(nil).Err))
}

// FieldDescriptions mocks base method.
func (m *Rows) FieldDescriptions() []pgconn.FieldDescription {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FieldDescriptions")
	ret0, _ := ret[0].([]pgconn.FieldDescription)
	return ret0
}

// FieldDescriptions indicates an expected call of FieldDescriptions.
func (mr *RowsMockRecorder) FieldDescriptions() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FieldDescriptions", reflect.TypeOf((*Rows)(nil).FieldDescriptions))
}

// Next mocks base method.
func (m *Rows) Next() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Next")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Next indicates an expected call of Next.
func (mr *RowsMockRecorder) Next() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Next", reflect.TypeOf((*Rows)(nil).Next))
}

// RawValues mocks base method.
func (m *Rows) RawValues() [][]byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RawValues")
	ret0, _ := ret[0].([][]byte)
	return ret0
}

// RawValues indicates an expected call of RawValues.
func (mr *RowsMockRecorder) RawValues() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RawValues", reflect.TypeOf((*Rows)(nil).RawValues))
}

// Scan mocks base method.
func (m *Rows) Scan(dest ...any) error {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range dest {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Scan", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Scan indicates an expected call of Scan.
func (mr *RowsMockRecorder) Scan(dest ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Scan", reflect.TypeOf((*Rows)(nil).Scan), dest...)
}

// Values mocks base method.
func (m *Rows) Values() ([]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Values")
	ret0, _ := ret[0].([]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Values indicates an expected call of Values.
func (mr *RowsMockRecorder) Values() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Values", reflect.TypeOf((*Rows)(nil).Values))
}
