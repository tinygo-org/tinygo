package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetPackageRoot(t *testing.T) {
	wantRoot := "github.com/my/root"

	gopath := os.TempDir()
	defer os.Unsetenv("GOPATH")
	os.Setenv("GOPATH", gopath)
	wd := filepath.Join(gopath, "src", wantRoot)
	err := os.MkdirAll(wd, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir(wd)
	if err != nil {
		t.Fatal(err)
	}

	gotRoot, err := getPackageRoot()
	if err != nil {
		t.Fatal(err)
	}
	if gotRoot != wantRoot {
		t.Fatalf("want %q got %q", wantRoot, gotRoot)
	}
}
