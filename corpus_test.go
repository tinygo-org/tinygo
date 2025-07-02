package main

import (
	"flag"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"golang.org/x/tools/go/buildutil"
	yaml "gopkg.in/yaml.v2"
)

/*
This contains code from https://github.com/dgryski/tinygo-test-corpus

The MIT License (MIT)

Copyright (c) 2020 Damian Gryski <damian@gryski.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.

*/

var corpus = flag.String("corpus", "", "path to test corpus")

func TestCorpus(t *testing.T) {
	t.Parallel()
	if *corpus == "" {
		t.Skip()
	}

	var target string
	if *testTarget != "" {
		target = *testTarget
	}
	isWASI := strings.HasPrefix(target, "wasi")

	repos, err := loadRepos(*corpus)
	if err != nil {
		t.Fatalf("loading corpus: %v", err)
	}

	for _, repo := range repos {
		repo := repo
		name := repo.Repo
		if repo.Tags != "" {
			name += "(" + strings.ReplaceAll(repo.Tags, " ", "-") + ")"
		}
		if repo.Version != "" {
			name += "@" + repo.Version
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if isWASI && repo.SkipWASI {
				t.Skip("skip wasi")
			}
			if repo.Slow && testing.Short() {
				t.Skip("slow test")
			}

			var wg sync.WaitGroup
			defer wg.Wait()
			out := ioLogger(t, &wg)
			defer out.Close()

			dir := t.TempDir()
			cmd := exec.Command("go", "mod", "init", "github.com/tinygo/tinygo-corpus-test")
			cmd.Dir = dir
			cmd.Stdout, cmd.Stderr = out, out
			err := cmd.Run()
			if err != nil {
				t.Errorf("failed to init: %s", err.Error())
				return
			}

			var ver string
			if repo.Version != "" {
				ver = "@" + repo.Version
			}
			cmd = exec.Command("go", "get", "-t", "-d", repo.Repo+"/..."+ver)
			cmd.Dir = dir
			cmd.Stdout, cmd.Stderr = out, out
			err = cmd.Run()
			if err != nil {
				t.Errorf("failed to get: %s", err.Error())
				return
			}

			doTest := func(t *testing.T, path string) {
				var wg sync.WaitGroup
				defer wg.Wait()
				out := ioLogger(t, &wg)
				defer out.Close()

				opts := optionsFromTarget(target, sema)
				opts.Directory = dir
				var tags buildutil.TagsFlag
				tags.Set(repo.Tags)
				opts.Tags = []string(tags)
				opts.TestConfig.Verbose = testing.Verbose()

				passed, err := Test(path, out, out, &opts, "")
				if err != nil {
					t.Errorf("test error: %v", err)
				}
				if !passed {
					t.Error("test failed")
				}
			}
			if len(repo.Subdirs) == 0 {
				doTest(t, repo.Repo)
				return
			}

			for _, dir := range repo.Subdirs {
				dir := dir
				t.Run(dir.Pkg, func(t *testing.T) {
					t.Parallel()

					if isWASI && dir.SkipWASI {
						t.Skip("skip wasi")
					}
					if dir.Slow && testing.Short() {
						t.Skip("slow test")
					}

					doTest(t, repo.Repo+"/"+dir.Pkg)
				})
			}
		})
	}
}

type T struct {
	Repo     string
	Tags     string
	Subdirs  []Subdir
	SkipWASI bool
	Slow     bool
	Version  string
}

type Subdir struct {
	Pkg      string
	SkipWASI bool
	Slow     bool
}

func loadRepos(f string) ([]T, error) {

	yf, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}

	var repos []T
	err = yaml.Unmarshal(yf, &repos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}
