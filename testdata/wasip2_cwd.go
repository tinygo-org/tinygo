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
		if _, err := os.ReadFile("filesystem.txt"); !os.IsNotExist(err) {
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
		if err := os.Chdir(".."); !os.IsNotExist(err) {
			panic("unmapped parent directory should deny Chdir")
		}
		if cwd, err := os.Getwd(); err != nil || cwd != initialCWD {
			panic("failed Chdir should not change cwd")
		}
		return
	}
	chdir := os.Getenv("EXPECT_CHDIR")
	if chdir == "" {
		chdir = ".."
	}
	if err := os.Chdir(chdir); err != nil {
		panic(err)
	}
	changed, err := os.Getwd()
	if err != nil || path.Clean(changed) != path.Join(initialCWD, chdir) {
		panic("unexpected cwd after Chdir: " + changed)
	}
	for i := 0; i < 3; i++ {
		runtime.GC()
		if err := os.Chdir("."); err != nil {
			panic(err)
		}
		if cwd, err := os.Getwd(); err != nil || cwd != changed {
			panic("unexpected cwd after repeated Chdir")
		}
	}
	if initialCWD != os.Getenv("EXPECT_CWD") {
		panic("Getwd result changed after Chdir")
	}
}
