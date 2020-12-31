package sh

import (
	"bytes"
	"io"
	"os"
	"syscall"
)

type Capture struct {
	f        *os.File
	releaseF *os.File
	w        *os.File
	out      chan string
}

// CaptureStdout buffers os.Stdout.
func CaptureStdout() *Capture {
	return captureFile(os.Stdout, os.NewFile(uintptr(syscall.Stdout), "/dev/stdout"))
}

// CaptureStderr buffers os.Stderr.
func CaptureStderr() *Capture {
	return captureFile(os.Stderr, os.NewFile(uintptr(syscall.Stderr), "/dev/stderr"))
}

func captureFile(f *os.File, releaseF *os.File) *Capture {
	c := &Capture{
		f:        f,
		releaseF: releaseF,
	}
	r, w, _ := os.Pipe()
	c.w = w
	*f = *w

	c.out = make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		c.out <- buf.String()
	}()

	return c
}

// Release reverts the changes to the captured output.
func (c *Capture) Release() {
	if c.releaseF == nil {
		return
	}

	*c.f = *c.releaseF
	c.w.Close()
	c.releaseF = nil
}

// Output releases the captured file and returns the output.
func (c *Capture) Output() string {
	c.Release()
	return <-c.out
}
