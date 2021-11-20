// +build darwin linux,!baremetal freebsd,!baremetal

package os_test

import (
	. "os"
	"testing"
)

func randomName(t *testing.T, pattern string) string {
	f, err := CreateTemp("", "TestMkdir")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	name := f.Name()
	f.Close()
	err = Remove(name)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	return name
}

func TestMkdir(t *testing.T) {
	dir := randomName(t, "TestMkdir")
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
	f := randomName(t, "TestRemove")

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
		// TODO: make this pass.
		if !IsNotExist(err) {
			t.Logf("TestRemove: TODO: expected IsNotExist(err) true, got false; err %q", err.Error())
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
