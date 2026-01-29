package output

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

// MockFormatter implements Formatter for testing
// It provides:
//   - Buffer capture for all output
//   - Call recording for verification
//   - Configurable error injection
//   - Easy access to captured output
type MockFormatter struct {
	mu sync.RWMutex

	// Output buffers
	buffer  *bytes.Buffer
	writer  io.Writer
	verbose bool
	style   Style

	// Calls records all method calls for verification
	Calls []MockFormatterCall

	// Captured structured data
	Tables    []MockTable
	Lists     [][]string
	KeyValues []map[string]string
	Objects   []interface{}
	Sections  []string
	Messages  []MockMessage

	// Error injection
	ObjectError error
}

// MockFormatterCall records a single method call
type MockFormatterCall struct {
	Method string
	Args   []interface{}
}

// MockTable records a Table() call
type MockTable struct {
	Headers []string
	Rows    [][]string
}

// MockMessage records a message output
type MockMessage struct {
	Type    string // info, success, warning, error, debug, progress
	Message string
}

// NewMockFormatter creates a new mock formatter
func NewMockFormatter() *MockFormatter {
	buf := &bytes.Buffer{}
	return &MockFormatter{
		buffer:    buf,
		writer:    buf,
		style:     StylePlain,
		Calls:     make([]MockFormatterCall, 0),
		Tables:    make([]MockTable, 0),
		Lists:     make([][]string, 0),
		KeyValues: make([]map[string]string, 0),
		Objects:   make([]interface{}, 0),
		Sections:  make([]string, 0),
		Messages:  make([]MockMessage, 0),
	}
}

// =============================================================================
// Message Output
// =============================================================================

func (m *MockFormatter) Info(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Info", message)
	m.Messages = append(m.Messages, MockMessage{Type: "info", Message: message})
	fmt.Fprintf(m.writer, "[INFO] %s\n", message)
}

func (m *MockFormatter) Success(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Success", message)
	m.Messages = append(m.Messages, MockMessage{Type: "success", Message: message})
	fmt.Fprintf(m.writer, "[OK] %s\n", message)
}

func (m *MockFormatter) Warning(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Warning", message)
	m.Messages = append(m.Messages, MockMessage{Type: "warning", Message: message})
	fmt.Fprintf(m.writer, "[WARN] %s\n", message)
}

func (m *MockFormatter) Error(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Error", message)
	m.Messages = append(m.Messages, MockMessage{Type: "error", Message: message})
	fmt.Fprintf(m.writer, "[ERROR] %s\n", message)
}

func (m *MockFormatter) Debug(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Debug", message)
	if m.verbose {
		m.Messages = append(m.Messages, MockMessage{Type: "debug", Message: message})
		fmt.Fprintf(m.writer, "[DEBUG] %s\n", message)
	}
}

func (m *MockFormatter) Progress(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Progress", message)
	m.Messages = append(m.Messages, MockMessage{Type: "progress", Message: message})
	fmt.Fprintf(m.writer, "-> %s\n", message)
}

func (m *MockFormatter) Step(number int, total int, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Step", number, total, message)
	m.Messages = append(m.Messages, MockMessage{
		Type:    "step",
		Message: fmt.Sprintf("[%d/%d] %s", number, total, message),
	})
	fmt.Fprintf(m.writer, "[%d/%d] %s\n", number, total, message)
}

// =============================================================================
// Structured Output
// =============================================================================

func (m *MockFormatter) Table(headers []string, rows [][]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Table", headers, rows)
	m.Tables = append(m.Tables, MockTable{Headers: headers, Rows: rows})

	// Write simple table representation
	fmt.Fprintln(m.writer, "TABLE:")
	fmt.Fprintf(m.writer, "  Headers: %v\n", headers)
	for i, row := range rows {
		fmt.Fprintf(m.writer, "  Row %d: %v\n", i, row)
	}
}

func (m *MockFormatter) List(items []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("List", items)
	m.Lists = append(m.Lists, items)

	fmt.Fprintln(m.writer, "LIST:")
	for _, item := range items {
		fmt.Fprintf(m.writer, "  - %s\n", item)
	}
}

func (m *MockFormatter) KeyValue(pairs map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("KeyValue", pairs)
	m.KeyValues = append(m.KeyValues, pairs)

	fmt.Fprintln(m.writer, "KEY-VALUE:")
	for k, v := range pairs {
		fmt.Fprintf(m.writer, "  %s: %s\n", k, v)
	}
}

func (m *MockFormatter) Object(v interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Object", v)

	if m.ObjectError != nil {
		return m.ObjectError
	}

	m.Objects = append(m.Objects, v)
	fmt.Fprintf(m.writer, "OBJECT: %+v\n", v)
	return nil
}

// =============================================================================
// Sections and Grouping
// =============================================================================

func (m *MockFormatter) Section(title string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Section", title)
	m.Sections = append(m.Sections, title)
	fmt.Fprintf(m.writer, "\n=== %s ===\n", title)
}

func (m *MockFormatter) Subsection(title string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Subsection", title)
	fmt.Fprintf(m.writer, "\n--- %s ---\n", title)
}

func (m *MockFormatter) Separator() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Separator")
	fmt.Fprintln(m.writer, "---")
}

func (m *MockFormatter) NewLine() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("NewLine")
	fmt.Fprintln(m.writer)
}

// =============================================================================
// Raw Output
// =============================================================================

func (m *MockFormatter) Print(text string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Print", text)
	fmt.Fprint(m.writer, text)
}

func (m *MockFormatter) Printf(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Printf", format, args)
	fmt.Fprintf(m.writer, format, args...)
}

func (m *MockFormatter) Println(text string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("Println", text)
	fmt.Fprintln(m.writer, text)
}

// =============================================================================
// Configuration
// =============================================================================

func (m *MockFormatter) SetWriter(w io.Writer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("SetWriter", w)
	m.writer = w
}

func (m *MockFormatter) SetVerbose(verbose bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recordCall("SetVerbose", verbose)
	m.verbose = verbose
}

func (m *MockFormatter) GetStyle() Style {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.style
}

// =============================================================================
// Test Helper Methods
// =============================================================================

// Reset clears all state for a fresh test
func (m *MockFormatter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.buffer = &bytes.Buffer{}
	m.writer = m.buffer
	m.Calls = make([]MockFormatterCall, 0)
	m.Tables = make([]MockTable, 0)
	m.Lists = make([][]string, 0)
	m.KeyValues = make([]map[string]string, 0)
	m.Objects = make([]interface{}, 0)
	m.Sections = make([]string, 0)
	m.Messages = make([]MockMessage, 0)
	m.ObjectError = nil
}

// Output returns all captured output as a string
func (m *MockFormatter) Output() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.buffer.String()
}

// Contains checks if the output contains a substring
func (m *MockFormatter) Contains(substr string) bool {
	return bytes.Contains(m.buffer.Bytes(), []byte(substr))
}

// CallCount returns the number of times a method was called
func (m *MockFormatter) CallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, call := range m.Calls {
		if call.Method == method {
			count++
		}
	}
	return count
}

// GetCalls returns all calls to a specific method
func (m *MockFormatter) GetCalls(method string) []MockFormatterCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var calls []MockFormatterCall
	for _, call := range m.Calls {
		if call.Method == method {
			calls = append(calls, call)
		}
	}
	return calls
}

// LastCall returns the last call made
func (m *MockFormatter) LastCall() *MockFormatterCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.Calls) == 0 {
		return nil
	}
	return &m.Calls[len(m.Calls)-1]
}

// GetMessagesByType returns all messages of a specific type
func (m *MockFormatter) GetMessagesByType(msgType string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var messages []string
	for _, msg := range m.Messages {
		if msg.Type == msgType {
			messages = append(messages, msg.Message)
		}
	}
	return messages
}

// HasError checks if any error messages were output
func (m *MockFormatter) HasError() bool {
	return len(m.GetMessagesByType("error")) > 0
}

// HasWarning checks if any warning messages were output
func (m *MockFormatter) HasWarning() bool {
	return len(m.GetMessagesByType("warning")) > 0
}

// SetStyle sets the style (for testing style-dependent code)
func (m *MockFormatter) SetStyle(style Style) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.style = style
}

// recordCall records a method call (internal, must be called with lock held)
func (m *MockFormatter) recordCall(method string, args ...interface{}) {
	m.Calls = append(m.Calls, MockFormatterCall{
		Method: method,
		Args:   args,
	})
}

// Ensure MockFormatter implements Formatter
var _ Formatter = (*MockFormatter)(nil)
