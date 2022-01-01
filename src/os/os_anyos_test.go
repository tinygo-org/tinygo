// +build windows darwin linux,!baremetal

package os_test

import (
	. "os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func randomName() string {
	// fastrand() does not seem available here, so fake it
	ns := time.Now().Nanosecond()
	pid := Getpid()
	return strconv.FormatUint(uint64(ns^pid), 10)
}

func TestMkdir(t *testing.T) {
	dir := TempDir() + "/TestMkdir" + randomName()
	Remove(dir)
	err := Mkdir(dir, 0755)
	defer Remove(dir)
	if err != nil {
		t.Errorf("Mkdir(%s, 0755) returned %v", dir, err)
	}
	// tests the "directory" branch of Remove
	err = Remove(dir)
	if err != nil {
		t.Errorf("Remove(%s) returned %v", dir, err)
	}
}

func TestStatBadDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Log("TODO: TestStatBadDir fails on Windows, skipping")
		return
	}
	dir := TempDir()
	badDir := filepath.Join(dir, "not-exist/really-not-exist")
	_, err := Stat(badDir)
	// TODO: PathError moved to io/fs in go 1.16; fix next line once we drop go 1.15 support.
	if pe, ok := err.(*PathError); !ok || !IsNotExist(err) || pe.Path != badDir {
		t.Errorf("Mkdir error = %#v; want PathError for path %q satisifying IsNotExist", err, badDir)
	}
}

func writeFile(t *testing.T, fname string, flag int, text string) string {
	f, err := OpenFile(fname, flag, 0666)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	n, err := f.WriteString(text)
	if err != nil {
		t.Fatalf("WriteString: %d, %v", n, err)
	}
	f.Close()
	data, err := ReadFile(f.Name())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	return string(data)
}

func TestRemove(t *testing.T) {
	f := TempDir() + "/TestRemove" + randomName()

	err := Remove(f)
	if err == nil {
		t.Errorf("TestRemove: remove of nonexistent file did not fail")
	} else {
		// FIXME: once we drop go 1.15, switch this to fs.PathError
		if pe, ok := err.(*PathError); !ok {
			t.Errorf("TestRemove: expected PathError, got err %q", err.Error())
		} else {
			if pe.Path != f {
				t.Errorf("TestRemove: PathError returned path %q, expected %q", pe.Path, f)
			}
		}
		if !IsNotExist(err) {
			t.Errorf("TestRemove: expected IsNotExist(err) true, got false; err %q", err.Error())
		}
	}

	s := writeFile(t, f, O_CREATE|O_TRUNC|O_RDWR, "new")
	if s != "new" {
		t.Fatalf("writeFile: have %q want %q", s, "new")
	}
	// tests the "file" branch of Remove
	err = Remove(f)
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
}
