package output

import (
	"bytes"
	"errors"
	"testing"
)

// =============================================================================
// Mock Implementation Tests
// =============================================================================

func TestMockFormatter_ImplementsInterface(t *testing.T) {
	// Compile-time check that MockFormatter implements Formatter
	var _ Formatter = (*MockFormatter)(nil)
}

func TestMockFormatter_MessageOutput(t *testing.T) {
	mock := NewMockFormatter()

	// Test all message types
	mock.Info("info message")
	mock.Success("success message")
	mock.Warning("warning message")
	mock.Error("error message")
	mock.Progress("progress message")
	mock.Step(1, 5, "step message")

	// Verify messages were recorded
	messages := mock.Messages
	if len(messages) != 6 {
		t.Errorf("Expected 6 messages, got %d", len(messages))
	}

	// Check specific types
	if len(mock.GetMessagesByType("info")) != 1 {
		t.Error("Expected 1 info message")
	}
	if len(mock.GetMessagesByType("success")) != 1 {
		t.Error("Expected 1 success message")
	}
	if len(mock.GetMessagesByType("error")) != 1 {
		t.Error("Expected 1 error message")
	}

	// Check output buffer
	output := mock.Output()
	if !mock.Contains("[INFO] info message") {
		t.Errorf("Output missing info message: %s", output)
	}
	if !mock.Contains("[OK] success message") {
		t.Errorf("Output missing success message: %s", output)
	}
}

func TestMockFormatter_Debug_VerboseOff(t *testing.T) {
	mock := NewMockFormatter()
	// verbose is false by default

	mock.Debug("debug message")

	// Debug should not appear in output when verbose is off
	if mock.Contains("[DEBUG]") {
		t.Error("Debug message should not appear when verbose is off")
	}
	if len(mock.GetMessagesByType("debug")) != 0 {
		t.Error("Debug message should not be recorded when verbose is off")
	}
}

func TestMockFormatter_Debug_VerboseOn(t *testing.T) {
	mock := NewMockFormatter()
	mock.SetVerbose(true)

	mock.Debug("debug message")

	if !mock.Contains("[DEBUG] debug message") {
		t.Error("Debug message should appear when verbose is on")
	}
	if len(mock.GetMessagesByType("debug")) != 1 {
		t.Error("Debug message should be recorded when verbose is on")
	}
}

func TestMockFormatter_Table(t *testing.T) {
	mock := NewMockFormatter()

	headers := []string{"NAME", "STATUS", "AGE"}
	rows := [][]string{
		{"project1", "running", "5d"},
		{"project2", "stopped", "2d"},
	}

	mock.Table(headers, rows)

	if len(mock.Tables) != 1 {
		t.Fatalf("Expected 1 table, got %d", len(mock.Tables))
	}

	table := mock.Tables[0]
	if len(table.Headers) != 3 {
		t.Errorf("Expected 3 headers, got %d", len(table.Headers))
	}
	if len(table.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(table.Rows))
	}

	// Verify call recorded
	if mock.CallCount("Table") != 1 {
		t.Errorf("CallCount(Table) = %d, want 1", mock.CallCount("Table"))
	}
}

func TestMockFormatter_List(t *testing.T) {
	mock := NewMockFormatter()

	items := []string{"item1", "item2", "item3"}
	mock.List(items)

	if len(mock.Lists) != 1 {
		t.Fatalf("Expected 1 list, got %d", len(mock.Lists))
	}

	if len(mock.Lists[0]) != 3 {
		t.Errorf("Expected 3 items, got %d", len(mock.Lists[0]))
	}
}

func TestMockFormatter_KeyValue(t *testing.T) {
	mock := NewMockFormatter()

	pairs := map[string]string{
		"name":   "test",
		"status": "running",
	}
	mock.KeyValue(pairs)

	if len(mock.KeyValues) != 1 {
		t.Fatalf("Expected 1 key-value set, got %d", len(mock.KeyValues))
	}

	if mock.KeyValues[0]["name"] != "test" {
		t.Errorf("KeyValue name = %q, want 'test'", mock.KeyValues[0]["name"])
	}
}

func TestMockFormatter_Object(t *testing.T) {
	mock := NewMockFormatter()

	obj := struct {
		Name  string
		Value int
	}{"test", 42}

	err := mock.Object(obj)
	if err != nil {
		t.Fatalf("Object() error = %v", err)
	}

	if len(mock.Objects) != 1 {
		t.Fatalf("Expected 1 object, got %d", len(mock.Objects))
	}
}

func TestMockFormatter_Object_Error(t *testing.T) {
	mock := NewMockFormatter()
	mock.ObjectError = errors.New("serialization failed")

	err := mock.Object(struct{}{})
	if err == nil {
		t.Error("Object() expected error, got nil")
	}
	if err.Error() != "serialization failed" {
		t.Errorf("Object() error = %v, want 'serialization failed'", err)
	}
}

func TestMockFormatter_Sections(t *testing.T) {
	mock := NewMockFormatter()

	mock.Section("Main Section")
	mock.Subsection("Sub Section")
	mock.Separator()
	mock.NewLine()

	if len(mock.Sections) != 1 {
		t.Errorf("Expected 1 section, got %d", len(mock.Sections))
	}
	if mock.Sections[0] != "Main Section" {
		t.Errorf("Section title = %q, want 'Main Section'", mock.Sections[0])
	}

	output := mock.Output()
	if !mock.Contains("=== Main Section ===") {
		t.Errorf("Missing section header in output: %s", output)
	}
	if !mock.Contains("--- Sub Section ---") {
		t.Errorf("Missing subsection header in output: %s", output)
	}
}

func TestMockFormatter_RawOutput(t *testing.T) {
	mock := NewMockFormatter()

	mock.Print("raw")
	mock.Printf(" formatted %d", 42)
	mock.Println(" line")

	output := mock.Output()
	expected := "raw formatted 42 line\n"
	if output != expected {
		t.Errorf("Output = %q, want %q", output, expected)
	}
}

func TestMockFormatter_SetWriter(t *testing.T) {
	mock := NewMockFormatter()
	custom := &bytes.Buffer{}

	mock.SetWriter(custom)
	mock.Info("test message")

	// Should write to custom writer
	if !bytes.Contains(custom.Bytes(), []byte("test message")) {
		t.Error("Output should go to custom writer")
	}
}

func TestMockFormatter_GetStyle(t *testing.T) {
	mock := NewMockFormatter()

	if mock.GetStyle() != StylePlain {
		t.Errorf("GetStyle() = %v, want StylePlain", mock.GetStyle())
	}

	mock.SetStyle(StyleColored)
	if mock.GetStyle() != StyleColored {
		t.Errorf("GetStyle() = %v, want StyleColored", mock.GetStyle())
	}
}

func TestMockFormatter_Reset(t *testing.T) {
	mock := NewMockFormatter()

	// Add various state
	mock.Info("test")
	mock.Table([]string{"H"}, [][]string{{"R"}})
	mock.List([]string{"item"})
	mock.ObjectError = errors.New("error")

	// Reset
	mock.Reset()

	// Verify all state cleared
	if len(mock.Calls) != 0 {
		t.Error("Reset() should clear Calls")
	}
	if len(mock.Tables) != 0 {
		t.Error("Reset() should clear Tables")
	}
	if len(mock.Lists) != 0 {
		t.Error("Reset() should clear Lists")
	}
	if len(mock.Messages) != 0 {
		t.Error("Reset() should clear Messages")
	}
	if mock.Output() != "" {
		t.Error("Reset() should clear output buffer")
	}
	if mock.ObjectError != nil {
		t.Error("Reset() should clear ObjectError")
	}
}

func TestMockFormatter_CallTracking(t *testing.T) {
	mock := NewMockFormatter()

	mock.Info("msg1")
	mock.Info("msg2")
	mock.Success("success")

	// Test CallCount
	if mock.CallCount("Info") != 2 {
		t.Errorf("CallCount(Info) = %d, want 2", mock.CallCount("Info"))
	}
	if mock.CallCount("Success") != 1 {
		t.Errorf("CallCount(Success) = %d, want 1", mock.CallCount("Success"))
	}

	// Test GetCalls
	infoCalls := mock.GetCalls("Info")
	if len(infoCalls) != 2 {
		t.Errorf("GetCalls(Info) returned %d calls, want 2", len(infoCalls))
	}

	// Test LastCall
	lastCall := mock.LastCall()
	if lastCall == nil {
		t.Fatal("LastCall() returned nil")
	}
	if lastCall.Method != "Success" {
		t.Errorf("LastCall().Method = %q, want 'Success'", lastCall.Method)
	}
}

func TestMockFormatter_HasError_HasWarning(t *testing.T) {
	mock := NewMockFormatter()

	if mock.HasError() {
		t.Error("HasError() should be false initially")
	}
	if mock.HasWarning() {
		t.Error("HasWarning() should be false initially")
	}

	mock.Error("something went wrong")
	if !mock.HasError() {
		t.Error("HasError() should be true after Error()")
	}

	mock.Warning("be careful")
	if !mock.HasWarning() {
		t.Error("HasWarning() should be true after Warning()")
	}
}

// =============================================================================
// Interface Compliance Tests
// =============================================================================

func TestFormatter_Interface_Mock(t *testing.T) {
	// Use mock through the interface
	var formatter Formatter = NewMockFormatter()

	// Test all interface methods work
	formatter.Info("info")
	formatter.Success("success")
	formatter.Warning("warning")
	formatter.Error("error")
	formatter.Debug("debug")
	formatter.Progress("progress")
	formatter.Step(1, 2, "step")

	formatter.Table([]string{"H"}, [][]string{{"R"}})
	formatter.List([]string{"item"})
	formatter.KeyValue(map[string]string{"k": "v"})
	formatter.Object(struct{}{})

	formatter.Section("section")
	formatter.Subsection("subsection")
	formatter.Separator()
	formatter.NewLine()

	formatter.Print("print")
	formatter.Printf("printf %d", 1)
	formatter.Println("println")

	formatter.SetWriter(&bytes.Buffer{})
	formatter.SetVerbose(true)
	_ = formatter.GetStyle()

	// If we get here without panic, the interface is fully implemented
}

// TestFormatter_Swappability verifies different formatter implementations
// can be swapped through the interface
func TestFormatter_Swappability(t *testing.T) {
	// Function that works with any Formatter
	testOutput := func(f Formatter, t *testing.T) {
		f.Section("Test Section")
		f.Info("Test message")
		f.Success("Complete")

		style := f.GetStyle()
		if style == "" {
			t.Error("GetStyle() returned empty string")
		}
	}

	t.Run("MockFormatter", func(t *testing.T) {
		mock := NewMockFormatter()
		testOutput(mock, t)

		// Verify mock captured the calls
		if mock.CallCount("Section") != 1 {
			t.Error("Mock should capture Section call")
		}
	})

	t.Run("PlainFormatter", func(t *testing.T) {
		buf := &bytes.Buffer{}
		plain := NewPlainFormatter(FormatterConfig{
			Writer: buf,
			Style:  StylePlain,
		}, PlainIcons())
		testOutput(plain, t)

		// Verify output was written
		if buf.Len() == 0 {
			t.Error("PlainFormatter should write output")
		}
	})

	// ColoredFormatter could be tested here too if needed
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestMockFormatter_ConcurrentAccess(t *testing.T) {
	mock := NewMockFormatter()

	done := make(chan bool, 10)

	// Multiple goroutines writing
	for i := 0; i < 5; i++ {
		go func(id int) {
			mock.Info("concurrent info")
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		go func(id int) {
			mock.Table([]string{"H"}, [][]string{{"R"}})
			done <- true
		}(i)
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify calls were recorded
	if len(mock.Calls) == 0 {
		t.Error("No calls recorded during concurrent access")
	}
}
