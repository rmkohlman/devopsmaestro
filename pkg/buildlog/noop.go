package buildlog

import "io"

// noopLogger is returned by New when Options.Enabled is false. It satisfies
// the BuildLogger interface but performs no disk I/O.
type noopLogger struct{}

func (n *noopLogger) Open(sessionID string) error       { return nil }
func (n *noopLogger) Writer(workspace string) io.Writer { return io.Discard }
func (n *noopLogger) Path() string                      { return "" }
func (n *noopLogger) Close() error                      { return nil }
