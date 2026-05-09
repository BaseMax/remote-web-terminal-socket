// Package ptybridge provides a platform-independent interface to a
// pseudo-terminal (PTY on Unix, ConPTY on Windows).
package ptybridge

// Bridge abstracts a running shell attached to a pseudo-terminal.
// Call Start to obtain one; call Close when done.
type Bridge interface {
	// Read reads output bytes produced by the shell.
	Read(p []byte) (int, error)
	// Write sends input bytes to the shell.
	Write(p []byte) (int, error)
	// Resize informs the terminal emulator of a new window size.
	Resize(cols, rows uint16) error
	// Close kills the shell process and releases all resources.
	Close() error
}

// Start launches the given shell binary inside a pseudo-terminal and
// returns a Bridge connected to it.
// shell must be an absolute path or a name resolvable via PATH.
func Start(shell string) (Bridge, error) {
	return start(shell)
}
