//go:build windows

package ptybridge

import (
	"github.com/UserExistsError/conpty"
)

type winBridge struct {
	cp *conpty.ConPty
}

func start(shell string) (Bridge, error) {
	// ConPtyDimensions sets a sane initial size; the client will send a
	// proper resize message immediately after connecting.
	cp, err := conpty.Start(shell, conpty.ConPtyDimensions(220, 50))
	if err != nil {
		return nil, err
	}
	return &winBridge{cp: cp}, nil
}

func (b *winBridge) Read(p []byte) (int, error)  { return b.cp.Read(p) }
func (b *winBridge) Write(p []byte) (int, error) { return b.cp.Write(p) }

func (b *winBridge) Resize(cols, rows uint16) error {
	return b.cp.Resize(int(cols), int(rows))
}

func (b *winBridge) Close() error {
	return b.cp.Close()
}
