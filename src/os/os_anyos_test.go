//go:build windows || darwin || (linux && !baremetal)

package os_test

import (
	"io/fs"
	"os"
	. "os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

var dot = []string{
	"dir.go",
	"env.go",
	"errors.go",
	"file.go",
	"os_test.go",
	"types.go",
	"stat_darwin.go",
	"stat_linux.go",
}

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
		t.Log("TODO: TestStatBadDir: IsNotExist fails on Windows, skipping")
		return
	}
	dir := TempDir()
	badDir := filepath.Join(dir, "not-exist/really-not-exist")
	_, err := Stat(badDir)
	if pe, ok := err.(*fs.PathError); !ok || !IsNotExist(err) || pe.Path != badDir {
		t.Errorf("Mkdir error = %#v; want PathError for path %q satisifying IsNotExist", err, badDir)
	}
}

func equal(name1, name2 string) (r bool) {
	switch runtime.GOOS {
	case "windows":
		r = strings.ToLower(name1) == strings.ToLower(name2)
	default:
		r = name1 == name2
	}
	return
}

func TestFstat(t *testing.T) {
	if runtime.GOARCH == "386" || runtime.GOARCH == "arm" {
		t.Log("TODO: implement fstat for 386 and arm")
		return
	}
	sfname := "TestFstat"
	path := TempDir() + "/" + sfname
	payload := writeFile(t, path, O_CREATE|O_TRUNC|O_RDWR, "Hello")
	defer Remove(path)

	file, err1 := Open(path)
	if err1 != nil {
		t.Fatal("open failed:", err1)
	}
	defer file.Close()
	dir, err2 := file.Stat()
	if err2 != nil {
		t.Fatal("fstat failed:", err2)
	}
	if !equal(sfname, dir.Name()) {
		t.Error("name should be ", sfname, "; is", dir.Name())
	}
	filesize := len(payload)
	if dir.Size() != int64(filesize) {
		t.Error("size should be", filesize, "; is", dir.Size())
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
		if pe, ok := err.(*fs.PathError); !ok {
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

// chtmpdir changes the working directory to a new temporary directory and
// provides a cleanup function.
func chtmpdir(t *testing.T) func() {
	oldwd, err := Getwd()
	if err != nil {
		t.Fatalf("chtmpdir: %v", err)
	}
	d, err := MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("chtmpdir: %v", err)
	}
	if err := Chdir(d); err != nil {
		t.Fatalf("chtmpdir: %v", err)
	}
	return func() {
		if err := Chdir(oldwd); err != nil {
			t.Fatalf("chtmpdir: %v", err)
		}
		RemoveAll(d)
	}
}

func TestRename(t *testing.T) {
	// TODO: use t.TempDir()
	from, to := TempDir()+"/"+"TestRename-from", TempDir()+"/"+"TestRename-to"

	file, err := Create(from)
	defer Remove(from) // TODO: switch to t.Tempdir, remove this line
	if err != nil {
		t.Fatalf("open %q failed: %v", from, err)
	}
	defer Remove(to) // TODO: switch to t.Tempdir, remove this line
	if err = file.Close(); err != nil {
		t.Errorf("close %q failed: %v", from, err)
	}
	err = Rename(from, to)
	if err != nil {
		t.Fatalf("rename %q, %q failed: %v", to, from, err)
	}
	_, err = Stat(to)
	if err != nil {
		t.Errorf("stat %q failed: %v", to, err)
	}
}

func TestRenameOverwriteDest(t *testing.T) {
	from, to := TempDir()+"/"+"TestRenameOverwrite-from", TempDir()+"/"+"TestRenameOverwrite-to"

	toData := []byte("to")
	fromData := []byte("from")

	err := os.WriteFile(to, toData, 0777)
	defer Remove(to) // TODO: switch to t.Tempdir, remove this line
	if err != nil {
		t.Fatalf("write file %q failed: %v", to, err)
	}

	err = os.WriteFile(from, fromData, 0777)
	defer Remove(from) // TODO: switch to t.Tempdir, remove this line
	if err != nil {
		t.Fatalf("write file %q failed: %v", from, err)
	}
	err = Rename(from, to)
	if err != nil {
		t.Fatalf("rename %q, %q failed: %v", to, from, err)
	}

	_, err = Stat(from)
	if err == nil {
		t.Errorf("from file %q still exists", from)
	}
	if runtime.GOOS == "windows" {
		t.Log("TODO: TestRenameOverwriteDest: IsNotExist fails on Windows, skipping")
	} else if err != nil && !IsNotExist(err) {
		t.Fatalf("stat from: %v", err)
	}
	toFi, err := Stat(to)
	if err != nil {
		t.Fatalf("stat %q failed: %v", to, err)
	}
	if toFi.Size() != int64(len(fromData)) {
		t.Errorf(`"to" size = %d; want %d (old "from" size)`, toFi.Size(), len(fromData))
	}
}

func TestRenameFailed(t *testing.T) {
	from, to := TempDir()+"/"+"RenameFailed-from", TempDir()+"/"+"RenameFailed-to"

	err := Rename(from, to)
	switch err := err.(type) {
	case *LinkError:
		if err.Op != "rename" {
			t.Errorf("rename %q, %q: err.Op: want %q, got %q", from, to, "rename", err.Op)
		}
		if err.Old != from {
			t.Errorf("rename %q, %q: err.Old: want %q, got %q", from, to, from, err.Old)
		}
		if err.New != to {
			t.Errorf("rename %q, %q: err.New: want %q, got %q", from, to, to, err.New)
		}
	case nil:
		t.Errorf("rename %q, %q: expected error, got nil", from, to)
	default:
		t.Errorf("rename %q, %q: expected %T, got %T %v", from, to, new(LinkError), err, err)
	}
}

func TestUserHomeDir(t *testing.T) {
	dir, err := UserHomeDir()
	if dir == "" && err == nil {
		t.Fatal("UserHomeDir returned an empty string but no error")
	}
	if err != nil {
		t.Logf("UserHomeDir failed: %v", err)
		return
	}
	fi, err := Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !fi.IsDir() {
		t.Fatalf("dir %s is not directory; type = %v", dir, fi.Mode())
	}
}

func TestDirFS(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Log("TODO: implement Readdir for Windows")
		return
	}
	if runtime.GOARCH == "386" || runtime.GOARCH == "arm" {
		t.Log("TODO: implement seek for 386 and arm")
		return
	}
	if isWASI {
		t.Log("TODO: allow foo/bar/. as synonym for path foo/bar on wasi?")
		return
	}
	if err := fstest.TestFS(DirFS("./testdata/dirfs"), "a", "b", "dir/x"); err != nil {
		t.Fatal(err)
	}

	// Test that Open does not accept backslash as separator.
	d := DirFS(".")
	_, err := d.Open(`testdata\dirfs`)
	if err == nil {
		t.Fatalf(`Open testdata\dirfs succeeded`)
	}
}

func TestDirFSPathsValid(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Log("skipping on Windows")
		return
	}
	if isWASI {
		t.Log("skipping on wasi because it fails on wasi on windows")
		return
	}

	// TODO: switch back to t.TempDir once it's implemented
	d, err := MkdirTemp("", "TestDirFSPathsValid")
	if err != nil {
		t.Fatal(err)
	}
	defer Remove(d)
	if err := WriteFile(filepath.Join(d, "control.txt"), []byte(string("Hello, world!")), 0644); err != nil {
		t.Fatal(err)
	}
	defer Remove(filepath.Join(d, "control.txt"))
	if err := WriteFile(filepath.Join(d, `e:xperi\ment.txt`), []byte(string("Hello, colon and backslash!")), 0644); err != nil {
		t.Fatal(err)
	}
	defer Remove(filepath.Join(d, `e:xperi\ment.txt`))

	fsys := DirFS(d)
	err = fs.WalkDir(fsys, ".", func(path string, e fs.DirEntry, err error) error {
		if fs.ValidPath(e.Name()) {
			t.Logf("%q ok", e.Name())
		} else {
			t.Errorf("%q INVALID", e.Name())
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
