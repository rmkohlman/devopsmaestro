package render

import (
	"fmt"
	"os"
)

// Blank outputs an empty line to the default writer.
func Blank() error {
	w := GetWriter()
	_, err := fmt.Fprintln(w)
	return err
}

// Infof outputs a formatted info message.
func Infof(format string, args ...any) error {
	return Info(fmt.Sprintf(format, args...))
}

// Successf outputs a formatted success message.
func Successf(format string, args ...any) error {
	return Success(fmt.Sprintf(format, args...))
}

// Warningf outputs a formatted warning message.
func Warningf(format string, args ...any) error {
	return Warning(fmt.Sprintf(format, args...))
}

// Errorf outputs a formatted error message.
func Errorf(format string, args ...any) error {
	return Error(fmt.Sprintf(format, args...))
}

// Progressf outputs a formatted progress message.
func Progressf(format string, args ...any) error {
	return Progress(fmt.Sprintf(format, args...))
}

// InfoToStderr outputs an info message to os.Stderr.
func InfoToStderr(content string) error {
	return MsgTo(os.Stderr, "", Message{Level: LevelInfo, Content: content})
}

// WarningToStderr outputs a warning message to os.Stderr.
func WarningToStderr(content string) error {
	return MsgTo(os.Stderr, "", Message{Level: LevelWarning, Content: content})
}

// ErrorToStderr outputs an error message to os.Stderr.
func ErrorToStderr(content string) error {
	return MsgTo(os.Stderr, "", Message{Level: LevelError, Content: content})
}
