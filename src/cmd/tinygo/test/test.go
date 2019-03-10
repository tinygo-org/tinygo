package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func RunTests() error {
	_, err := findTestFiles()
	if err != nil {
		return err
	}

	return nil
}

func findTestFiles() ([]string, error) {
	var testFiles []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), "_test.go") {
			fmt.Println(path)
			testFiles = append(testFiles, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return testFiles, nil
}