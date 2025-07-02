package main

import (
	_ "github.com/tinygo-org/tinygo/testdata/errors/non-existing-package"
	_ "github.com/tinygo-org/tinygo/testdata/errors/non-existing-package-2"
)

func main() {
}

// ERROR: loader-nopackage.go:4:2: no required module provides package github.com/tinygo-org/tinygo/testdata/errors/non-existing-package; to add it:
// ERROR: 	go get github.com/tinygo-org/tinygo/testdata/errors/non-existing-package
// ERROR: loader-nopackage.go:5:2: no required module provides package github.com/tinygo-org/tinygo/testdata/errors/non-existing-package-2; to add it:
// ERROR: 	go get github.com/tinygo-org/tinygo/testdata/errors/non-existing-package-2
