package main

import (
	"os"
	"path"
	"runtime"
)

var initialCWD string
var unmappedCWD = os.Getenv("EXPECT_UNMAPPED_CWD") == "true"

func init() {
	var err error
	initialCWD, err = os.Getwd()
	if err != nil || initialCWD != os.Getenv("EXPECT_CWD") {
		panic("unexpected initial cwd: " + initialCWD)
	}
	if unmappedCWD {
		if _, err := os.ReadFile("filesystem.txt"); !os.IsPermission(err) {
			panic("unmapped cwd should deny relative file access")
		}
		if data, err := os.ReadFile("/workspace/testdata/filesystem.txt"); err != nil || len(data) == 0 {
			panic("unmapped cwd should not prevent absolute file access")
		}
		return
	}
	if data, err := os.ReadFile("filesystem.txt"); err != nil || len(data) == 0 {
		panic("relative read during init failed")
	}
}

func main() {
	runtime.GC()
	if unmappedCWD {
		if err := os.Chdir(".."); !os.IsPermission(err) {
			panic("unmapped parent directory should deny Chdir")
		}
		if cwd, err := os.Getwd(); err != nil || cwd != initialCWD {
			panic("failed Chdir should not change cwd")
		}
		return
	}
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	changed, err := os.Getwd()
	if err != nil || path.Clean(changed) != path.Dir(path.Clean(initialCWD)) {
		panic("unexpected cwd after Chdir: " + changed)
	}
	if initialCWD != os.Getenv("EXPECT_CWD") {
		panic("Getwd result changed after Chdir")
	}
}
