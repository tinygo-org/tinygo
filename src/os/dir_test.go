//go:build darwin || (linux && !baremetal && !js && !wasip1 && !wasip2 && !386 && !arm)

package os_test

import (
	. "os"
	"testing"
)

func testReaddirnames(dir string, contents []string, t *testing.T) {
	file, err := Open(dir)
	if err != nil {
		t.Fatalf("open %q failed: %v", dir, err)
	}
	defer file.Close()
	s, err2 := file.Readdirnames(-1)
	if err2 != nil {
		t.Fatalf("Readdirnames %q failed: %v", dir, err2)
	}
	for _, m := range contents {
		found := false
		for _, n := range s {
			if n == "." || n == ".." {
				t.Errorf("got %q in directory", n)
			}
			if !equal(m, n) {
				continue
			}
			if found {
				t.Error("present twice:", m)
			}
			found = true
		}
		if !found {
			t.Error("could not find", m)
		}
	}
	if s == nil {
		t.Error("Readdirnames returned nil instead of empty slice")
	}
}

func testReaddir(dir string, contents []string, t *testing.T) {
	file, err := Open(dir)
	if err != nil {
		t.Fatalf("open %q failed: %v", dir, err)
	}
	defer file.Close()
	s, err2 := file.Readdir(-1)
	if err2 != nil {
		t.Fatalf("Readdir %q failed: %v", dir, err2)
	}
	for _, m := range contents {
		found := false
		for _, n := range s {
			if n.Name() == "." || n.Name() == ".." {
				t.Errorf("got %q in directory", n.Name())
			}
			if !equal(m, n.Name()) {
				continue
			}
			if found {
				t.Error("present twice:", m)
			}
			found = true
		}
		if !found {
			t.Error("could not find", m)
		}
	}
	if s == nil {
		t.Error("Readdir returned nil instead of empty slice")
	}
}

func testReadDir(dir string, contents []string, t *testing.T) {
	file, err := Open(dir)
	if err != nil {
		t.Fatalf("open %q failed: %v", dir, err)
	}
	defer file.Close()
	s, err2 := file.ReadDir(-1)
	if err2 != nil {
		t.Fatalf("ReadDir %q failed: %v", dir, err2)
	}
	for _, m := range contents {
		found := false
		for _, n := range s {
			if n.Name() == "." || n.Name() == ".." {
				t.Errorf("got %q in directory", n)
			}
			if !equal(m, n.Name()) {
				continue
			}
			if found {
				t.Error("present twice:", m)
			}
			found = true
			lstat, err := Lstat(dir + "/" + m)
			if err != nil {
				t.Fatal(err)
			}
			if n.IsDir() != lstat.IsDir() {
				t.Errorf("%s: IsDir=%v, want %v", m, n.IsDir(), lstat.IsDir())
			}
			if n.Type() != lstat.Mode().Type() {
				t.Errorf("%s: IsDir=%v, want %v", m, n.Type(), lstat.Mode().Type())
			}
			info, err := n.Info()
			if err != nil {
				t.Errorf("%s: Info: %v", m, err)
				continue
			}
			if !SameFile(info, lstat) {
				t.Errorf("%s: Info: SameFile(info, lstat) = false", m)
			}
		}
		if !found {
			t.Error("could not find", m)
		}
	}
	if s == nil {
		t.Error("ReadDir returned nil instead of empty slice")
	}
}

func TestFileReaddirnames(t *testing.T) {
	testReaddirnames(".", dot, t)
	testReaddirnames(TempDir(), nil, t)
}

func TestFileReaddir(t *testing.T) {
	testReaddir(".", dot, t)
	testReaddir(TempDir(), nil, t)
}

func TestFileReadDir(t *testing.T) {
	testReadDir(".", dot, t)
	testReadDir(TempDir(), nil, t)
}

func TestReadDir(t *testing.T) {
	dirname := "rumpelstilzchen"
	_, err := ReadDir(dirname)
	if err == nil {
		t.Fatalf("ReadDir %s: error expected, none found", dirname)
	}

	dirname = "testdata"
	list, err := ReadDir(dirname)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", dirname, err)
	}

	foundFile := false
	foundSubDir := false
	for _, dir := range list {
		switch {
		case !dir.IsDir() && dir.Name() == "hello":
			foundFile = true
		case dir.IsDir() && dir.Name() == "issue37161":
			foundSubDir = true
		}
	}
	if !foundFile {
		t.Fatalf("ReadDir %s: hello file not found", dirname)
	}
	if !foundSubDir {
		t.Fatalf("ReadDir %s: testdata directory not found", dirname)
	}
}

// TestReadNonDir just verifies that opening a non-directory returns an error.
func TestReadNonDir(t *testing.T) {
	// Use filename of this source file; it is known to exist, and go tests run in source tree.
	dirname := "dir_test.go"
	_, err := ReadDir(dirname)
	if err == nil {
		t.Fatalf("ReadDir %s: error on non-dir expected, none found", dirname)
	}
}
