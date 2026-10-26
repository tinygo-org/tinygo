package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	closed  bool
	// cleanup removes what the program left behind.
	cleanup func()
}

// exited reports whether the program ended and all its output was read.
func (p *pipePort) exited() bool {
	return p.closed && len(p.pending) == 0
}

func (p *pipePort) Read(b []byte) (int, error) {
	if len(p.pending) == 0 {
		select {
		case data, ok := <-p.out:
			if !ok {
				p.closed = true
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
	if p.cleanup != nil {
		p.cleanup()
	}
	return nil
}

// buildFlags returns the flags that change how the programs are built.
func buildFlags(o options) []string {
	var flags []string
	if o.noInput {
		flags = append(flags, "-tags", "qbs_noinput")
	}
	if o.scheduler != "" {
		flags = append(flags, "-scheduler", o.scheduler)
	}
	if o.gc != "" {
		flags = append(flags, "-gc", o.gc)
	}
	return flags
}

// startEmulator runs the program of a class in the emulator of the target.
// Output and input go through the returned link.
func startEmulator(o options, dir string) (*link, error) {
	if o.pty {
		return startEmulatorPTY(o, dir)
	}
	args := append([]string{"run", "-target", o.target}, buildFlags(o)...)
	cmd := exec.Command(o.tinygo, append(args, dir)...)
	cmd.Dir = dir
	return connect(o, cmd, nil)
}

// startEmulatorPTY builds the program and starts the emulator itself under a
// pseudo terminal. tinygo run gives the emulator a pipe, and an emulator such
// as simavr then holds back its output until the program ends.
func startEmulatorPTY(o options, dir string) (*link, error) {
	info, err := exec.Command(o.tinygo, "info", "-json", "-target", o.target).Output()
	if err != nil {
		return nil, fmt.Errorf("tinygo info failed: %v", err)
	}
	var spec struct {
		Target struct {
			Emulator string `json:"emulator"`
		} `json:"target"`
	}
	if err := json.Unmarshal(info, &spec); err != nil || spec.Target.Emulator == "" {
		return nil, fmt.Errorf("the target %s has no emulator", o.target)
	}
	tmp, err := os.MkdirTemp("", "qbs")
	if err != nil {
		return nil, err
	}
	prog := filepath.Join(tmp, "prog.elf")
	build := append([]string{"build", "-target", o.target, "-o", prog}, buildFlags(o)...)
	if out, err := exec.Command(o.tinygo, append(build, dir)...).CombinedOutput(); err != nil {
		os.RemoveAll(tmp)
		return nil, fmt.Errorf("build failed: %s", out)
	}
	words := strings.Fields(strings.ReplaceAll(spec.Target.Emulator, "{}", prog))
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("script", append([]string{"-q", "/dev/null"}, words...)...)
	} else {
		cmd = exec.Command("script", "-qfec", shellJoin(words), "/dev/null")
	}
	cmd.Dir = dir
	cleanup := func() {
		// script starts the emulator in its own session, so stop it by the
		// path of the program, which is unique to this run.
		exec.Command("pkill", "-f", prog).Run()
		os.RemoveAll(tmp)
	}
	l, err := connect(o, cmd, cleanup)
	if err != nil {
		cleanup()
	}
	return l, err
}

// connect starts the command and connects its input and output to a link.
func connect(o options, cmd *exec.Cmd, cleanup func()) (*link, error) {
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
	return &link{port: &pipePort{cmd: cmd, in: in, out: out, cleanup: cleanup}}, nil
}

// shellJoin quotes the arguments for a shell command line.
func shellJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		quoted[i] = "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
	}
	return strings.Join(quoted, " ")
}

// emulatorFailure returns a short reason when the program could not run
// because the emulator is missing or the program does not fit on the target.
func emulatorFailure(log []byte) string {
	text := string(log)
	if i := strings.Index(text, "executable file not found"); i >= 0 {
		return "emulator not installed: " + strings.TrimSpace(text[max(0, i-60):i+len("executable file not found")])
	}
	return fitFailure(text)
}

// fitFailure returns a reason, with the line of the error, when the build
// output says that the program does not fit on the target. The markers are
// kept narrow so that other link errors still fail.
func fitFailure(text string) string {
	for _, line := range strings.Split(text, "\n") {
		for _, marker := range []string{"overflowed", "will not fit", "is out of range"} {
			if strings.Contains(line, marker) {
				line = strings.TrimSpace(line)
				if len(line) > 110 {
					line = line[:107] + "..."
				}
				return "program does not fit on this target: " + line
			}
		}
	}
	return ""
}
