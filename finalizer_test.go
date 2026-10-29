package main

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/builder"
)

func TestBoehmFinalizerInterior(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		target, scheduler string
	}{
		{"wasm", "asyncify"},
		{"wasm", "none"},
		{"", "tasks"},
		{"", "threads"},
		{"", "none"},
	} {
		if test.target == "" && runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
			continue
		}
		t.Run(test.target+"/"+test.scheduler, func(t *testing.T) {
			t.Parallel()
			options := optionsFromTarget(test.target, sema)
			options.GC = "boehm"
			options.Scheduler = test.scheduler
			options.Tags = append(options.Tags, "runtime_asserts")
			options.GlobalValues = map[string]map[string]string{
				"main": {"finalizerTestMode": "interior"},
			}
			runTest("finalizer.go", options, t, nil, nil)
		})
	}
}

func TestBoehmFinalizerInteriorInvalid(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("host finalizers need Linux or Darwin")
	}
	options := optionsFromTarget("", sema)
	options.GC = "boehm"
	config, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"interior-pointer", "interior-large"} {
		t.Run(name, func(t *testing.T) {
			output := &bytes.Buffer{}
			_, err := buildAndRun("testdata/finalizer.go", config, output, []string{name}, nil, time.Minute, func(cmd *exec.Cmd, result *builder.BuildResult) error {
				cmd.Stdout = nil
				cmd.Stderr = nil
				data, err := cmd.CombinedOutput()
				output.Write(data)
				return err
			})
			if err == nil || !strings.Contains(output.String(), "pointer not at beginning of allocated block") {
				t.Fatalf("interior registration result: %v\n%s", err, output.String())
			}
		})
	}
}

func TestFinalizerRunnerRegistration(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		target, gc, scheduler string
	}{
		{"wasm", "boehm", "asyncify"},
		{"wasm", "conservative", "asyncify"},
		{"", "boehm", "tasks"},
		{"", "boehm", "threads"},
	} {
		if test.target == "" && runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
			continue
		}
		t.Run(test.target+"/"+test.gc+"/"+test.scheduler, func(t *testing.T) {
			t.Parallel()
			options := optionsFromTarget(test.target, sema)
			options.GC = test.gc
			options.Scheduler = test.scheduler
			options.GlobalValues = map[string]map[string]string{
				"main": {"finalizerIdleMode": "registration"},
			}
			runTest("finalizeridle.go", options, t, nil, nil)
		})
	}
}
