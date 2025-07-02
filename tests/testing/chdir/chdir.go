package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
)

/*
Test that this program is 'run' in expected directory. 'run' with expected
working-directory in 'EXPECT_DIR' environment variable' with{,out} a -C
argument.
*/
func main() {
	expectDir := os.Getenv("EXPECT_DIR")
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		cwd = filepath.ToSlash(cwd)
	}
	if cwd != expectDir {
		log.Fatalf("expected:\"%v\" != os.Getwd():\"%v\"", expectDir, cwd)
	}
}
