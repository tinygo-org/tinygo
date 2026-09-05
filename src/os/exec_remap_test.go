//go:build (linux || darwin) && !baremetal && !tinygo.wasm && !nintendoswitch

package os_test

import (
	"errors"
	"fmt"
	. "os"
	"syscall"
	"testing"
)

func TestForkExecFileRemapping(t *testing.T) {
	if _, err := Stat("/bin/bash"); err != nil {
		t.Skip("file remapping checks need /bin/bash for descriptors above 9")
	}
	for _, name := range []string{"cycle", "repeated", "closed-source", "identity", "sparse"} {
		t.Run(name, func(t *testing.T) {
			a, err := CreateTemp(t.TempDir(), "source-a")
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			b, err := CreateTemp(t.TempDir(), "source-b")
			if err != nil {
				t.Fatal(err)
			}
			defer b.Close()
			out, err := CreateTemp(t.TempDir(), "output")
			if err != nil {
				t.Fatal(err)
			}
			defer out.Close()
			for i, f := range []*File{a, b} {
				if _, err := f.WriteString([]string{"A\n", "B\n"}[i]); err != nil {
					t.Fatal(err)
				}
				if _, err := f.Seek(0, 0); err != nil {
					t.Fatal(err)
				}
			}
			x, y := int(a.Fd()), int(b.Fd())
			if x < 3 || y <= x {
				t.Fatalf("unexpected source descriptors %d, %d", x, y)
			}
			files := make([]*File, y+3)
			files[1], files[2] = out, Stderr
			targets := []int{x, y}
			switch name {
			case "cycle":
				files[x], files[y] = b, a
			case "repeated":
				files[x], files[y], files[y+1] = b, a, a
				targets = []int{x, y, y + 1}
			case "closed-source":
				files[y] = a
				targets = []int{y}
			case "identity":
				files[x], files[y] = a, b
			case "sparse":
				files[y+2] = a
				targets = []int{y + 2}
			}
			for _, fd := range targets {
				for _, f := range []*File{a, b} {
					if _, err := f.Seek(0, 0); err != nil {
						t.Fatal(err)
					}
				}
				want := "A"
				if files[fd] == b {
					want = "B"
				}
				script := fmt.Sprintf("exec 0<&%d; IFS= read -r value; test \"$value\" = \"$1\"", fd)
				checkRemapChild(t, files, script, want, true)
			}
			nextfd := len(files)
			for _, f := range files {
				if f != nil && int(f.Fd()) >= nextfd {
					nextfd = int(f.Fd()) + 1
				}
			}
			for i, f := range files {
				if f == nil && i >= 3 {
					script := fmt.Sprintf("exec 2>/dev/null; exec 0<&%d", i)
					checkRemapChild(t, files, script, "", false)
				}
				if f != nil && int(f.Fd()) < i {
					script := fmt.Sprintf("exec 2>/dev/null; exec 0<&%d", nextfd)
					checkRemapChild(t, files, script, "", false)
					nextfd++
				}
			}
			for i, f := range []*File{a, b} {
				got := make([]byte, 1)
				_, err := f.ReadAt(got, 0)
				if err != nil || string(got) != []string{"A", "B"}[i] {
					t.Fatalf("parent source = %q, %v", got, err)
				}
			}
		})
	}
}

func checkRemapChild(t *testing.T, files []*File, script, want string, success bool) {
	t.Helper()
	proc, err := StartProcess("/bin/bash", []string{"bash", "-c", script, "bash", want}, &ProcAttr{Files: files})
	if err != nil {
		t.Fatalf("StartProcess for %q = %v", script, err)
	}
	state, err := proc.Wait()
	if err != nil || state.Success() != success {
		t.Fatalf("Wait for %q = %v, %v, want success %v", script, state, err, success)
	}
}

func TestForkExecInvalidRemapSource(t *testing.T) {
	f, err := Open(DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	files := make([]*File, int(f.Fd())+2)
	files[len(files)-1] = f
	files[0] = NewFile(1<<31, "invalid")
	proc, err := StartProcess("/bin/sh", []string{"sh", "-c", "exit 0"}, &ProcAttr{Files: files})
	if proc != nil {
		proc.Kill()
		proc.Wait()
		t.Fatal("StartProcess accepted an invalid descriptor")
	}
	if !errors.Is(err, syscall.EBADF) {
		t.Fatalf("StartProcess = %v, want EBADF", err)
	}
	if _, err := f.Stat(); err != nil {
		t.Fatalf("parent source stat = %v", err)
	}
}
