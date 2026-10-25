package main

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// conn is what the runner needs from a serial port or an emulated program.
type conn interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	Close() error
}

// pipePort talks to a program run by tinygo run. Its Read gives up after a
// short time without data, like a serial port with a read timeout.
type pipePort struct {
	cmd     *exec.Cmd
	in      io.WriteCloser
	out     chan []byte
	pending []byte
}

func (p *pipePort) Read(b []byte) (int, error) {
	if len(p.pending) == 0 {
		select {
		case data, ok := <-p.out:
			if !ok {
				time.Sleep(50 * time.Millisecond)
				return 0, nil
			}
			p.pending = data
		case <-time.After(50 * time.Millisecond):
			return 0, nil
		}
	}
	n := copy(b, p.pending)
	p.pending = p.pending[n:]
	return n, nil
}

func (p *pipePort) Write(b []byte) (int, error) {
	return p.in.Write(b)
}

func (p *pipePort) Close() error {
	p.in.Close()
	killGroup(p.cmd)
	p.cmd.Wait()
	return nil
}

// startEmulator runs the program of a class with tinygo run, which uses the
// emulator of the target. Output and input go through the returned link.
func startEmulator(o options, dir string) (*link, error) {
	cmd := exec.Command(o.tinygo, "run", "-target", o.target, dir)
	cmd.Dir = dir
	if o.pathPrepend != "" {
		cmd.Env = append(os.Environ(), "PATH="+o.pathPrepend+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
	setProcessGroup(cmd)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	// Compiler messages and the program's own output arrive on one stream.
	cmd.Stdout = w
	cmd.Stderr = w
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	w.Close()
	out := make(chan []byte, 64)
	go func() {
		defer close(out)
		for {
			buf := make([]byte, 4096)
			n, err := r.Read(buf)
			if n > 0 {
				out <- buf[:n]
			}
			if err != nil {
				return
			}
		}
	}()
	return &link{port: &pipePort{cmd: cmd, in: in, out: out}}, nil
}

// emulatorFailure returns a short reason when the emulator is missing.
func emulatorFailure(log []byte) string {
	text := string(log)
	if i := strings.Index(text, "executable file not found"); i >= 0 {
		return "emulator not installed: " + strings.TrimSpace(text[max(0, i-60):i+len("executable file not found")])
	}
	return ""
}
