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

// Test that a panic in a finalizer run inline by scheduler=none cannot be
// recovered by the code that triggered the GC.
func TestFinalizerPanic(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"finalizerpanic.go", "finalizerpanic-alloc.go"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := optionsFromTarget("cortex-m-qemu", sema)
			options.Scheduler = "none"
			emuCheck(t, options)
			config, err := builder.NewConfig(&options)
			if err != nil {
				t.Fatal(err)
			}

			output := &bytes.Buffer{}
			_, err = buildAndRun("testdata/"+name, config, output, nil, nil, time.Minute, func(cmd *exec.Cmd, result *builder.BuildResult) error {
				cmd.Stdout = nil
				cmd.Stderr = nil
				data, err := cmd.CombinedOutput()
				output.Write(data)
				return err
			})
			if err == nil {
				t.Fatal("program unexpectedly exited successfully")
			}
			if want := "panic: finalizer panic"; !strings.Contains(output.String(), want) {
				t.Fatalf("output does not contain %q:\n%s", want, output.String())
			}
			if strings.Contains(output.String(), "recovered:") || strings.Contains(output.String(), "finished") {
				t.Fatalf("finalizer panic did not stop the program:\n%s", output.String())
			}
		})
	}
}

func TestFinalizerPanicRecovery(t *testing.T) {
	t.Parallel()

	options := optionsFromTarget("cortex-m-qemu", sema)
	options.Scheduler = "none"
	emuCheck(t, options)
	config, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}

	output := &bytes.Buffer{}
	_, err = buildAndRun("testdata/finalizerpanic-recover.go", config, output, nil, nil, time.Minute, func(cmd *exec.Cmd, result *builder.BuildResult) error {
		cmd.Stdout = nil
		cmd.Stderr = nil
		data, err := cmd.CombinedOutput()
		output.Write(data)
		return err
	})
	if err != nil {
		t.Fatalf("program failed: %v\n%s", err, output.String())
	}
	for _, want := range []string{"finalizer recovered", "caller recovered"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("output does not contain %q:\n%s", want, output.String())
		}
	}
}
