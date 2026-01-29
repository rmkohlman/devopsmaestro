package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// baseFormatter provides common functionality for all formatters
type baseFormatter struct {
	writer  io.Writer
	verbose bool
	style   Style
	icons   Icons
	config  FormatterConfig
}

func newBaseFormatter(cfg FormatterConfig, icons Icons) baseFormatter {
	writer := cfg.Writer
	if writer == nil {
		writer = os.Stdout
	}
	return baseFormatter{
		writer:  writer,
		verbose: cfg.Verbose,
		style:   cfg.Style,
		icons:   icons,
		config:  cfg,
	}
}

func (b *baseFormatter) SetWriter(w io.Writer) {
	b.writer = w
}

func (b *baseFormatter) SetVerbose(verbose bool) {
	b.verbose = verbose
}

func (b *baseFormatter) GetStyle() Style {
	return b.style
}

func (b *baseFormatter) write(s string) {
	fmt.Fprint(b.writer, s)
}

func (b *baseFormatter) writeln(s string) {
	fmt.Fprintln(b.writer, s)
}

func (b *baseFormatter) writef(format string, args ...interface{}) {
	fmt.Fprintf(b.writer, format, args...)
}

// PlainFormatter outputs plain text without any formatting
type PlainFormatter struct {
	baseFormatter
}

// NewPlainFormatter creates a new plain text formatter
func NewPlainFormatter(cfg FormatterConfig, icons Icons) *PlainFormatter {
	return &PlainFormatter{
		baseFormatter: newBaseFormatter(cfg, PlainIcons()),
	}
}

func (f *PlainFormatter) Info(message string)    { f.writeln(message) }
func (f *PlainFormatter) Success(message string) { f.writeln("[OK] " + message) }
func (f *PlainFormatter) Warning(message string) { f.writeln("[WARN] " + message) }
func (f *PlainFormatter) Error(message string)   { f.writeln("[ERROR] " + message) }
func (f *PlainFormatter) Debug(message string) {
	if f.verbose {
		f.writeln("[DEBUG] " + message)
	}
}
func (f *PlainFormatter) Progress(message string) { f.writeln("-> " + message) }
func (f *PlainFormatter) Step(num, total int, message string) {
	f.writef("[%d/%d] %s\n", num, total, message)
}

func (f *PlainFormatter) Table(headers []string, rows [][]string) {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers
	for i, h := range headers {
		f.writef("%-*s  ", widths[i], h)
	}
	f.writeln("")

	// Print separator
	for i := range headers {
		f.write(strings.Repeat("-", widths[i]) + "  ")
	}
	f.writeln("")

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				f.writef("%-*s  ", widths[i], cell)
			}
		}
		f.writeln("")
	}
}

func (f *PlainFormatter) List(items []string) {
	for _, item := range items {
		f.writeln("  * " + item)
	}
}

func (f *PlainFormatter) KeyValue(pairs map[string]string) {
	for k, v := range pairs {
		f.writef("%s: %s\n", k, v)
	}
}

func (f *PlainFormatter) Object(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	f.writeln(string(data))
	return nil
}

func (f *PlainFormatter) Section(title string) {
	f.writeln("")
	f.writeln("== " + title + " ==")
}

func (f *PlainFormatter) Subsection(title string) {
	f.writeln("-- " + title)
}

func (f *PlainFormatter) Separator() {
	f.writeln(strings.Repeat("-", 40))
}

func (f *PlainFormatter) NewLine()                                  { f.writeln("") }
func (f *PlainFormatter) Print(text string)                         { f.write(text) }
func (f *PlainFormatter) Printf(format string, args ...interface{}) { f.writef(format, args...) }
func (f *PlainFormatter) Println(text string)                       { f.writeln(text) }

// JSONFormatter outputs JSON for machine consumption
type JSONFormatter struct {
	baseFormatter
	encoder *json.Encoder
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter(cfg FormatterConfig) *JSONFormatter {
	writer := cfg.Writer
	if writer == nil {
		writer = os.Stdout
	}
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return &JSONFormatter{
		baseFormatter: newBaseFormatter(cfg, PlainIcons()),
		encoder:       encoder,
	}
}

func (f *JSONFormatter) outputMessage(level, message string) {
	f.encoder.Encode(map[string]string{"level": level, "message": message})
}

func (f *JSONFormatter) Info(message string)    { f.outputMessage("info", message) }
func (f *JSONFormatter) Success(message string) { f.outputMessage("success", message) }
func (f *JSONFormatter) Warning(message string) { f.outputMessage("warning", message) }
func (f *JSONFormatter) Error(message string)   { f.outputMessage("error", message) }
func (f *JSONFormatter) Debug(message string) {
	if f.verbose {
		f.outputMessage("debug", message)
	}
}
func (f *JSONFormatter) Progress(message string) { f.outputMessage("progress", message) }
func (f *JSONFormatter) Step(num, total int, message string) {
	f.encoder.Encode(map[string]interface{}{
		"level":   "step",
		"step":    num,
		"total":   total,
		"message": message,
	})
}

func (f *JSONFormatter) Table(headers []string, rows [][]string) {
	data := make([]map[string]string, len(rows))
	for i, row := range rows {
		data[i] = make(map[string]string)
		for j, cell := range row {
			if j < len(headers) {
				data[i][headers[j]] = cell
			}
		}
	}
	f.encoder.Encode(data)
}

func (f *JSONFormatter) List(items []string)              { f.encoder.Encode(items) }
func (f *JSONFormatter) KeyValue(pairs map[string]string) { f.encoder.Encode(pairs) }
func (f *JSONFormatter) Object(v interface{}) error       { return f.encoder.Encode(v) }

func (f *JSONFormatter) Section(title string)                      {} // No-op for JSON
func (f *JSONFormatter) Subsection(title string)                   {} // No-op for JSON
func (f *JSONFormatter) Separator()                                {} // No-op for JSON
func (f *JSONFormatter) NewLine()                                  {} // No-op for JSON
func (f *JSONFormatter) Print(text string)                         {}
func (f *JSONFormatter) Printf(format string, args ...interface{}) {}
func (f *JSONFormatter) Println(text string)                       {}

// YAMLFormatter outputs YAML for structured data
type YAMLFormatter struct {
	baseFormatter
}

// NewYAMLFormatter creates a new YAML formatter
func NewYAMLFormatter(cfg FormatterConfig) *YAMLFormatter {
	return &YAMLFormatter{
		baseFormatter: newBaseFormatter(cfg, PlainIcons()),
	}
}

func (f *YAMLFormatter) outputMessage(level, message string) {
	data, _ := yaml.Marshal(map[string]string{"level": level, "message": message})
	f.write(string(data))
	f.writeln("---")
}

func (f *YAMLFormatter) Info(message string)    { f.outputMessage("info", message) }
func (f *YAMLFormatter) Success(message string) { f.outputMessage("success", message) }
func (f *YAMLFormatter) Warning(message string) { f.outputMessage("warning", message) }
func (f *YAMLFormatter) Error(message string)   { f.outputMessage("error", message) }
func (f *YAMLFormatter) Debug(message string) {
	if f.verbose {
		f.outputMessage("debug", message)
	}
}
func (f *YAMLFormatter) Progress(message string) { f.outputMessage("progress", message) }
func (f *YAMLFormatter) Step(num, total int, message string) {
	data, _ := yaml.Marshal(map[string]interface{}{
		"level":   "step",
		"step":    num,
		"total":   total,
		"message": message,
	})
	f.write(string(data))
	f.writeln("---")
}

func (f *YAMLFormatter) Table(headers []string, rows [][]string) {
	data := make([]map[string]string, len(rows))
	for i, row := range rows {
		data[i] = make(map[string]string)
		for j, cell := range row {
			if j < len(headers) {
				data[i][headers[j]] = cell
			}
		}
	}
	out, _ := yaml.Marshal(data)
	f.writeln(string(out))
}

func (f *YAMLFormatter) List(items []string) {
	out, _ := yaml.Marshal(items)
	f.writeln(string(out))
}

func (f *YAMLFormatter) KeyValue(pairs map[string]string) {
	out, _ := yaml.Marshal(pairs)
	f.writeln(string(out))
}

func (f *YAMLFormatter) Object(v interface{}) error {
	out, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	f.writeln(string(out))
	return nil
}

func (f *YAMLFormatter) Section(title string) {
	f.writeln("# " + title)
}

func (f *YAMLFormatter) Subsection(title string) {
	f.writeln("## " + title)
}

func (f *YAMLFormatter) Separator() {
	f.writeln("---")
}

func (f *YAMLFormatter) NewLine()                                  { f.writeln("") }
func (f *YAMLFormatter) Print(text string)                         { f.write(text) }
func (f *YAMLFormatter) Printf(format string, args ...interface{}) { f.writef(format, args...) }
func (f *YAMLFormatter) Println(text string)                       { f.writeln(text) }
