//go:build !windows

package ptybridge

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

type unixBridge struct {
	f   *os.File
	cmd *exec.Cmd
}

func start(shell string) (Bridge, error) {
	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	f, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}
	return &unixBridge{f: f, cmd: cmd}, nil
}

func (b *unixBridge) Read(p []byte) (int, error)  { return b.f.Read(p) }
func (b *unixBridge) Write(p []byte) (int, error) { return b.f.Write(p) }

func (b *unixBridge) Resize(cols, rows uint16) error {
	return pty.Setsize(b.f, &pty.Winsize{Cols: cols, Rows: rows})
}

func (b *unixBridge) Close() error {
	err := b.f.Close()
	if b.cmd.Process != nil {
		b.cmd.Process.Kill() //nolint:errcheck
	}
	b.cmd.Wait() //nolint:errcheck
	return err
}
