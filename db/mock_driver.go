package db

import (
	"context"
	"errors"
	"sync"
)

// MockDriver implements the Driver interface for testing.
// It records calls and allows setting up expectations.
type MockDriver struct {
	mu sync.Mutex

	// ConnectFunc is called when Connect() is invoked
	ConnectFunc func() error

	// CloseFunc is called when Close() is invoked
	CloseFunc func() error

	// PingFunc is called when Ping() is invoked
	PingFunc func() error

	// ExecuteFunc is called when Execute() is invoked
	ExecuteFunc func(query string, args ...interface{}) (Result, error)

	// QueryRowFunc is called when QueryRow() is invoked
	QueryRowFunc func(query string, args ...interface{}) Row

	// QueryFunc is called when Query() is invoked
	QueryFunc func(query string, args ...interface{}) (Rows, error)

	// BeginFunc is called when Begin() is invoked
	BeginFunc func() (Transaction, error)

	// TypeValue is returned by Type()
	TypeValue DriverType

	// DSNValue is returned by DSN()
	DSNValue string

	// MigrationDSNValue is returned by MigrationDSN()
	MigrationDSNValue string

	// Calls records all method calls for verification
	Calls []MockCall
}

// MockCall represents a recorded method call
type MockCall struct {
	Method string
	Args   []interface{}
}

// NewMockDriver creates a new mock driver with default implementations
func NewMockDriver() *MockDriver {
	return &MockDriver{
		TypeValue:         DriverMemory,
		DSNValue:          "mock://test",
		MigrationDSNValue: "mock://test/migrate",
		ConnectFunc:       func() error { return nil },
		CloseFunc:         func() error { return nil },
		PingFunc:          func() error { return nil },
		ExecuteFunc: func(query string, args ...interface{}) (Result, error) {
			return &MockResult{}, nil
		},
		QueryRowFunc: func(query string, args ...interface{}) Row {
			return &MockRow{}
		},
		QueryFunc: func(query string, args ...interface{}) (Rows, error) {
			return &MockRows{}, nil
		},
		BeginFunc: func() (Transaction, error) {
			return &MockTransaction{}, nil
		},
	}
}

func (d *MockDriver) recordCall(method string, args ...interface{}) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Calls = append(d.Calls, MockCall{Method: method, Args: args})
}

func (d *MockDriver) Connect() error {
	d.recordCall("Connect")
	if d.ConnectFunc != nil {
		return d.ConnectFunc()
	}
	return nil
}

func (d *MockDriver) Close() error {
	d.recordCall("Close")
	if d.CloseFunc != nil {
		return d.CloseFunc()
	}
	return nil
}

func (d *MockDriver) Ping() error {
	d.recordCall("Ping")
	if d.PingFunc != nil {
		return d.PingFunc()
	}
	return nil
}

func (d *MockDriver) Execute(query string, args ...interface{}) (Result, error) {
	d.recordCall("Execute", append([]interface{}{query}, args...)...)
	if d.ExecuteFunc != nil {
		return d.ExecuteFunc(query, args...)
	}
	return &MockResult{}, nil
}

func (d *MockDriver) ExecuteContext(ctx context.Context, query string, args ...interface{}) (Result, error) {
	d.recordCall("ExecuteContext", append([]interface{}{ctx, query}, args...)...)
	if d.ExecuteFunc != nil {
		return d.ExecuteFunc(query, args...)
	}
	return &MockResult{}, nil
}

func (d *MockDriver) QueryRow(query string, args ...interface{}) Row {
	d.recordCall("QueryRow", append([]interface{}{query}, args...)...)
	if d.QueryRowFunc != nil {
		return d.QueryRowFunc(query, args...)
	}
	return &MockRow{}
}

func (d *MockDriver) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	d.recordCall("QueryRowContext", append([]interface{}{ctx, query}, args...)...)
	if d.QueryRowFunc != nil {
		return d.QueryRowFunc(query, args...)
	}
	return &MockRow{}
}

func (d *MockDriver) Query(query string, args ...interface{}) (Rows, error) {
	d.recordCall("Query", append([]interface{}{query}, args...)...)
	if d.QueryFunc != nil {
		return d.QueryFunc(query, args...)
	}
	return &MockRows{}, nil
}

func (d *MockDriver) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	d.recordCall("QueryContext", append([]interface{}{ctx, query}, args...)...)
	if d.QueryFunc != nil {
		return d.QueryFunc(query, args...)
	}
	return &MockRows{}, nil
}

func (d *MockDriver) Begin() (Transaction, error) {
	d.recordCall("Begin")
	if d.BeginFunc != nil {
		return d.BeginFunc()
	}
	return &MockTransaction{}, nil
}

func (d *MockDriver) BeginContext(ctx context.Context) (Transaction, error) {
	d.recordCall("BeginContext", ctx)
	if d.BeginFunc != nil {
		return d.BeginFunc()
	}
	return &MockTransaction{}, nil
}

func (d *MockDriver) Type() DriverType {
	return d.TypeValue
}

func (d *MockDriver) DSN() string {
	return d.DSNValue
}

func (d *MockDriver) MigrationDSN() string {
	return d.MigrationDSNValue
}

// GetCalls returns all recorded calls
func (d *MockDriver) GetCalls() []MockCall {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]MockCall{}, d.Calls...)
}

// ResetCalls clears recorded calls
func (d *MockDriver) ResetCalls() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Calls = nil
}

// Ensure MockDriver implements Driver
var _ Driver = (*MockDriver)(nil)

// =============================================================================
// MockRow
// =============================================================================

// MockRow implements the Row interface for testing
type MockRow struct {
	ScanFunc func(dest ...interface{}) error
	ScanErr  error
	Values   []interface{}
}

func (r *MockRow) Scan(dest ...interface{}) error {
	if r.ScanFunc != nil {
		return r.ScanFunc(dest...)
	}
	if r.ScanErr != nil {
		return r.ScanErr
	}
	// Copy values to destinations if provided
	for i := 0; i < len(dest) && i < len(r.Values); i++ {
		if r.Values[i] != nil {
			// Simple value copy (works for basic types)
			switch d := dest[i].(type) {
			case *int:
				if v, ok := r.Values[i].(int); ok {
					*d = v
				}
			case *int64:
				if v, ok := r.Values[i].(int64); ok {
					*d = v
				}
			case *string:
				if v, ok := r.Values[i].(string); ok {
					*d = v
				}
			case *bool:
				if v, ok := r.Values[i].(bool); ok {
					*d = v
				}
			}
		}
	}
	return nil
}

var _ Row = (*MockRow)(nil)

// =============================================================================
// MockRows
// =============================================================================

// MockRows implements the Rows interface for testing
type MockRows struct {
	Data       [][]interface{}
	ColumnList []string
	index      int
	closed     bool
	err        error
}

func (r *MockRows) Next() bool {
	if r.closed || r.index >= len(r.Data) {
		return false
	}
	r.index++
	return r.index <= len(r.Data)
}

func (r *MockRows) Scan(dest ...interface{}) error {
	if r.closed {
		return errors.New("rows closed")
	}
	if r.index == 0 || r.index > len(r.Data) {
		return errors.New("no row to scan")
	}

	row := r.Data[r.index-1]
	for i := 0; i < len(dest) && i < len(row); i++ {
		if row[i] != nil {
			switch d := dest[i].(type) {
			case *int:
				if v, ok := row[i].(int); ok {
					*d = v
				}
			case *int64:
				if v, ok := row[i].(int64); ok {
					*d = v
				}
			case *string:
				if v, ok := row[i].(string); ok {
					*d = v
				}
			case *bool:
				if v, ok := row[i].(bool); ok {
					*d = v
				}
			}
		}
	}
	return nil
}

func (r *MockRows) Close() error {
	r.closed = true
	return nil
}

func (r *MockRows) Err() error {
	return r.err
}

func (r *MockRows) Columns() ([]string, error) {
	return r.ColumnList, nil
}

var _ Rows = (*MockRows)(nil)

// =============================================================================
// MockResult
// =============================================================================

// MockResult implements the Result interface for testing
type MockResult struct {
	LastID          int64
	AffectedRows    int64
	LastIDErr       error
	AffectedRowsErr error
}

func (r *MockResult) LastInsertId() (int64, error) {
	return r.LastID, r.LastIDErr
}

func (r *MockResult) RowsAffected() (int64, error) {
	return r.AffectedRows, r.AffectedRowsErr
}

var _ Result = (*MockResult)(nil)

// =============================================================================
// MockTransaction
// =============================================================================

// MockTransaction implements the Transaction interface for testing
type MockTransaction struct {
	ExecuteFunc  func(query string, args ...interface{}) (Result, error)
	QueryRowFunc func(query string, args ...interface{}) Row
	QueryFunc    func(query string, args ...interface{}) (Rows, error)
	CommitFunc   func() error
	RollbackFunc func() error
	Committed    bool
	RolledBack   bool
}

func (t *MockTransaction) Execute(query string, args ...interface{}) (Result, error) {
	if t.ExecuteFunc != nil {
		return t.ExecuteFunc(query, args...)
	}
	return &MockResult{}, nil
}

func (t *MockTransaction) QueryRow(query string, args ...interface{}) Row {
	if t.QueryRowFunc != nil {
		return t.QueryRowFunc(query, args...)
	}
	return &MockRow{}
}

func (t *MockTransaction) Query(query string, args ...interface{}) (Rows, error) {
	if t.QueryFunc != nil {
		return t.QueryFunc(query, args...)
	}
	return &MockRows{}, nil
}

func (t *MockTransaction) Commit() error {
	t.Committed = true
	if t.CommitFunc != nil {
		return t.CommitFunc()
	}
	return nil
}

func (t *MockTransaction) Rollback() error {
	t.RolledBack = true
	if t.RollbackFunc != nil {
		return t.RollbackFunc()
	}
	return nil
}

var _ Transaction = (*MockTransaction)(nil)

// =============================================================================
// MockDB (Legacy Database interface)
// =============================================================================

// MockDB implements the legacy Database interface for testing.
// Deprecated: Use MockDriver instead for new tests.
type MockDB struct {
	mu sync.Mutex

	// ConnectFunc is called when Connect() is invoked
	ConnectFunc func() error

	// CloseFunc is called when Close() is invoked
	CloseFunc func() error

	// ExecuteFunc is called when Execute() is invoked
	ExecuteFunc func(query string, args ...interface{}) (interface{}, error)

	// FetchOneFunc is called when FetchOne() is invoked
	FetchOneFunc func(query string, args ...interface{}) (interface{}, error)

	// FetchManyFunc is called when FetchMany() is invoked
	FetchManyFunc func(query string, args ...interface{}) (interface{}, error)

	// DSNValue is returned by DSN()
	DSNValue string

	// MigrationDSNValue is returned by MigrationDSN()
	MigrationDSNValue string

	// Calls records all method calls for verification
	Calls []MockCall
}

// NewMockDB creates a new mock database with default implementations
func NewMockDB() *MockDB {
	return &MockDB{
		ConnectFunc:       func() error { return nil },
		CloseFunc:         func() error { return nil },
		DSNValue:          "mock://test",
		MigrationDSNValue: "sqlite3://mock://test",
		ExecuteFunc: func(query string, args ...interface{}) (interface{}, error) {
			return nil, nil
		},
		FetchOneFunc: func(query string, args ...interface{}) (interface{}, error) {
			return nil, nil
		},
		FetchManyFunc: func(query string, args ...interface{}) (interface{}, error) {
			return nil, nil
		},
	}
}

func (d *MockDB) recordCall(method string, args ...interface{}) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Calls = append(d.Calls, MockCall{Method: method, Args: args})
}

func (d *MockDB) Connect() error {
	d.recordCall("Connect")
	if d.ConnectFunc != nil {
		return d.ConnectFunc()
	}
	return nil
}

func (d *MockDB) Close() error {
	d.recordCall("Close")
	if d.CloseFunc != nil {
		return d.CloseFunc()
	}
	return nil
}

func (d *MockDB) Execute(query string, args ...interface{}) (interface{}, error) {
	d.recordCall("Execute", append([]interface{}{query}, args...)...)
	if d.ExecuteFunc != nil {
		return d.ExecuteFunc(query, args...)
	}
	return nil, nil
}

func (d *MockDB) FetchOne(query string, args ...interface{}) (interface{}, error) {
	d.recordCall("FetchOne", append([]interface{}{query}, args...)...)
	if d.FetchOneFunc != nil {
		return d.FetchOneFunc(query, args...)
	}
	return nil, nil
}

func (d *MockDB) FetchMany(query string, args ...interface{}) (interface{}, error) {
	d.recordCall("FetchMany", append([]interface{}{query}, args...)...)
	if d.FetchManyFunc != nil {
		return d.FetchManyFunc(query, args...)
	}
	return nil, nil
}

func (d *MockDB) DSN() string {
	return d.DSNValue
}

func (d *MockDB) MigrationDSN() string {
	return d.MigrationDSNValue
}

// GetCalls returns all recorded calls
func (d *MockDB) GetCalls() []MockCall {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]MockCall{}, d.Calls...)
}

// ResetCalls clears recorded calls
func (d *MockDB) ResetCalls() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Calls = nil
}

// Ensure MockDB implements Database
var _ Database = (*MockDB)(nil)
