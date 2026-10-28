package os_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestWindowsReadDirectoryAfterNames(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"first", "second"} {
		if err := os.WriteFile(filepath.Join(root, name), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, mode := range []string{"ReadDir", "Readdir"} {
		t.Run(mode, func(t *testing.T) {
			f, err := os.Open(root)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			if names, err := f.Readdirnames(1); err != nil || len(names) != 1 {
				t.Fatalf("Readdirnames(1) = %v, %v", names, err)
			}
			var info os.FileInfo
			if mode == "ReadDir" {
				entries, err := f.ReadDir(1)
				if err != nil || len(entries) != 1 {
					t.Fatalf("ReadDir(1) = %v, %v", entries, err)
				}
				info, err = entries[0].Info()
				if err != nil {
					t.Fatal(err)
				}
			} else {
				infos, err := f.Readdir(1)
				if err != nil || len(infos) != 1 {
					t.Fatalf("Readdir(1) = %v, %v", infos, err)
				}
				info = infos[0]
			}
			want, err := os.Stat(filepath.Join(root, info.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if !os.SameFile(info, want) {
				t.Fatalf("SameFile returned false for %s", info.Name())
			}
		})
	}
}

func TestWindowsReadDirectory(t *testing.T) {
	root, err := os.MkdirTemp("", "TestWindowsReadDirectory")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	empty := filepath.Join(root, "empty")
	if err := os.Mkdir(empty, 0700); err != nil {
		t.Fatal(err)
	}
	want := []string{"empty", "file-\U0001f600.txt"}
	for i := 0; i < 80; i++ {
		want = append(want, fmt.Sprintf("%03d-%s", i, strings.Repeat("x", 80)))
	}
	for _, name := range want[1:] {
		if err := os.WriteFile(filepath.Join(root, name), []byte("hello"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	sort.Strings(want)
	readers := []struct {
		name string
		read func(*os.File, int) ([]string, error)
	}{
		{"Readdirnames", (*os.File).Readdirnames},
		{"ReadDir", func(f *os.File, n int) ([]string, error) {
			entries, err := f.ReadDir(n)
			names := make([]string, len(entries))
			for i, entry := range entries {
				names[i] = entry.Name()
			}
			return names, err
		}},
		{"Readdir", func(f *os.File, n int) ([]string, error) {
			infos, err := f.Readdir(n)
			names := make([]string, len(infos))
			for i, info := range infos {
				names[i] = info.Name()
			}
			return names, err
		}},
	}
	for _, reader := range readers {
		for _, n := range []int{1, 2, 0, -1} {
			for _, dir := range []string{root, empty} {
				t.Run(fmt.Sprintf("%s/%d/%s", reader.name, n, filepath.Base(dir)), func(t *testing.T) {
					f, err := os.Open(dir)
					if err != nil {
						t.Fatal(err)
					}
					defer f.Close()
					got := []string{}
					for {
						names, err := reader.read(f, n)
						if names == nil {
							t.Fatal("returned a nil slice")
						}
						if n > 0 && len(names) > n {
							t.Fatalf("read %d entries, want at most %d", len(names), n)
						}
						got = append(got, names...)
						if err != nil {
							if n <= 0 || err != io.EOF || len(names) != 0 {
								t.Fatalf("read returned %d entries and %v", len(names), err)
							}
							break
						}
						if n <= 0 {
							break
						}
						if len(names) == 0 || len(got) > len(want) {
							t.Fatal("read made no progress or repeated entries")
						}
					}
					sort.Strings(got)
					expected := want
					if dir == empty {
						expected = []string{}
					}
					if !reflect.DeepEqual(got, expected) {
						t.Fatalf("read %v, want %v", got, expected)
					}
					for i := 0; i < 2; i++ {
						names, err := reader.read(f, n)
						expectedErr := io.EOF
						if n <= 0 {
							expectedErr = nil
						}
						if names == nil || len(names) != 0 || err != expectedErr {
							t.Fatalf("read after end returned %v, %v; want [], %v", names, err, expectedErr)
						}
					}
					if err := f.Close(); err != nil {
						t.Fatal(err)
					}
					_, err = reader.read(f, n)
					if pe, ok := err.(*os.PathError); !ok || pe.Err != os.ErrClosed {
						t.Fatalf("read after Close returned %v, want ErrClosed", err)
					}
				})
			}
		}
	}
}
