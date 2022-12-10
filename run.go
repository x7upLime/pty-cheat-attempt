package pty

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// Start assigns a pseudo-terminal tty os.File to c.Stdin, c.Stdout,
// and c.Stderr, calls c.Start, and returns the File of the tty's
// corresponding pty.
//
// Starts the process in a new session and sets the controlling terminal.
func Start(cmd *exec.Cmd) (*moddedTerm, error) {
	fmt.Printf("=========x7Term=======================================\n")
	return StartWithSize(cmd, nil)
}

type moddedTerm struct {
	*os.File
}

func (x *moddedTerm) Write(b []byte) (int, error) {
	_, err := x.File.Write([]byte("something...\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "something went terribly wrong, wake the kids!\n")
		os.Exit(-1)
	}
	return x.File.Write(b)
}

// StartWithAttrs assigns a pseudo-terminal tty os.File to c.Stdin, c.Stdout,
// and c.Stderr, calls c.Start, and returns the File of the tty's
// corresponding pty.
//
// This will resize the pty to the specified size before starting the command if a size is provided.
// The `attrs` parameter overrides the one set in c.SysProcAttr.
//
// This should generally not be needed. Used in some edge cases where it is needed to create a pty
// without a controlling terminal.
func StartWithAttrs(c *exec.Cmd, sz *Winsize, attrs *syscall.SysProcAttr) (*moddedTerm, error) {
	var cheat moddedTerm
	pty, tty, err := Open()
	if err != nil {
		return nil, err
	}

	cheat = moddedTerm{tty}
	defer func() { _ = cheat.File.Close() }() // Best effort.

	if sz != nil {
		if err := Setsize(pty, sz); err != nil {
			_ = pty.Close() // Best effort.
			return nil, err
		}
	}
	if c.Stdout == nil {
		c.Stdout = &cheat
	}
	if c.Stderr == nil {
		c.Stderr = &cheat
	}
	if c.Stdin == nil {
		c.Stdin = &cheat
	}

	c.SysProcAttr = attrs

	if err := c.Start(); err != nil {
		_ = pty.Close() // Best effort.
		return nil, err
	}
	return &cheat, err
}
