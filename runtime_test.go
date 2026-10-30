package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/goenv"
)

// TestTimerStopResetRace checks that stopping or resetting a timer while its
// callback is running behaves correctly. It reaches into the runtime timer
// hooks via //go:linkname and relies on the threads scheduler, which is only
// used on these hosts.
func TestTimerStopResetRace(t *testing.T) {
	t.Parallel()

	switch runtime.GOOS {
	case "darwin", "linux":
	default:
		t.Skipf("host GOOS %s does not use the threads scheduler", runtime.GOOS)
	}
	runTest("timer_stop_reset_race.go", optionsFromTarget("", sema), t, nil, nil)
}

func TestGoexitCrash(t *testing.T) {
	t.Parallel()

	options := optionsFromTarget("", sema)
	config, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		want string
	}{
		{"main", "fatal error: all goroutines are asleep - deadlock!"},
		{"deadlock", "fatal error: all goroutines are asleep - deadlock!"},
		{"exit", "fatal error: all goroutines are asleep - deadlock!"},
		{"main-other", "fatal error: all goroutines are asleep - deadlock!"},
		{"in-panic", "fatal error: all goroutines are asleep - deadlock!"},
		{"panic", "panic: panic after Goexit"},
		{"recovered-panic", "fatal error: all goroutines are asleep - deadlock!"},
		{"recover-before-panic", "fatal error: all goroutines are asleep - deadlock!"},
		{"recover-before-panic-loop", "fatal error: all goroutines are asleep - deadlock!"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			_, err = buildAndRun("testdata/goexit.go", config, output, []string{tc.name}, nil, time.Minute, func(cmd *exec.Cmd, result builder.BuildResult) error {
				cmd.Stdout = nil
				cmd.Stderr = nil
				data, err := cmd.CombinedOutput()
				output.Write(data)
				return err
			})
			if err == nil {
				t.Fatal("program unexpectedly exited successfully")
			}
			if !strings.Contains(output.String(), tc.want) {
				t.Fatalf("output does not contain %q:\n%s", tc.want, output.String())
			}
		})
	}

	for _, tc := range []struct {
		name          string
		arg           string
		panicStrategy string
		want          string
	}{
		{"wasip1-deadlock", "deadlock", "", "fatal error: all goroutines are asleep - deadlock!"},
		{"wasip1-goexit-panic-trap", "defer", "trap", "defer ran"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			options := optionsFromTarget("wasip1", sema)
			options.PanicStrategy = tc.panicStrategy
			config, err := builder.NewConfig(&options)
			if err != nil {
				t.Fatal(err)
			}

			result, err := builder.Build("testdata/goexit.go", ".wasm", t.TempDir(), config)
			if err != nil {
				t.Fatal("failed to build binary:", err)
			}
			data, err := os.ReadFile(result.Binary)
			if err != nil {
				t.Fatal("failed to read binary:", err)
			}

			output := &bytes.Buffer{}
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfigInterpreter())
			defer r.Close(ctx)
			wasi_snapshot_preview1.MustInstantiate(ctx, r)
			moduleConfig := wazero.NewModuleConfig().
				WithArgs(result.Binary, tc.arg).
				WithStdout(output).
				WithStderr(output)
			_, err = r.InstantiateWithConfig(ctx, data, moduleConfig)
			if err == nil {
				t.Fatal("program unexpectedly exited successfully")
			}
			if !strings.Contains(output.String(), tc.want) {
				t.Fatalf("output does not contain %q:\n%s", tc.want, output.String())
			}
		})
	}
}

func TestRuntimeFatal(t *testing.T) {
	t.Parallel()

	options := optionsFromTarget("", sema)
	config, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}

	output := &bytes.Buffer{}
	_, err = buildAndRun("testdata/runtimefatal.go", config, output, nil, nil, time.Minute, func(cmd *exec.Cmd, result builder.BuildResult) error {
		cmd.Stdout = nil
		cmd.Stderr = nil
		data, err := cmd.CombinedOutput()
		output.Write(data)
		return err
	})
	if err == nil {
		t.Fatal("program unexpectedly exited successfully")
	}
	if want := "fatal error: sync: unlock of unlocked Mutex"; !strings.Contains(output.String(), want) {
		t.Fatalf("output does not contain %q:\n%s", want, output.String())
	}
	if strings.Contains(output.String(), "recovered:") {
		t.Fatalf("fatal runtime error was recovered:\n%s", output.String())
	}
}

// Test that the program can read stdin when the caller supplies it.
func TestStdin(t *testing.T) {
	t.Parallel()

	options := optionsFromTarget("", sema)
	config, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}

	output := &bytes.Buffer{}
	_, err = buildAndRun(TESTDATA+"/stdin.go", config, output, nil, nil, time.Minute, func(cmd *exec.Cmd, result builder.BuildResult) error {
		cmd.Stdin = strings.NewReader("hello\nworld\n")
		return cmd.Run()
	})
	if err != nil {
		t.Error("failed to run:", err)
	}
	checkOutput(t, TESTDATA+"/stdin.txt", output.Bytes())
}

func TestRuntimeGOROOT(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name        string
		environment []string
		want        string
	}{
		{"startup environment", []string{"GOROOT=/startup"}, "/startup"},
		{"empty startup environment", []string{"GOROOT="}, goenv.Get("GOROOT")},
	} {
		t.Run(test.name, func(t *testing.T) {
			options := optionsFromTarget(*testTarget, sema)
			config, err := builder.NewConfig(&options)
			if err != nil {
				t.Fatal(err)
			}

			output := &bytes.Buffer{}
			_, err = buildAndRun("testdata/runtimegoroot.go", config, output, nil, test.environment, time.Minute, func(cmd *exec.Cmd, _ builder.BuildResult) error {
				return cmd.Run()
			})
			if err != nil {
				t.Fatal(err)
			}
			want := test.want + "\n" + test.want + "\n"
			if output.String() != want {
				t.Fatalf("unexpected output:\n%s\nwant:\n%s", output.String(), want)
			}
		})
	}
}

func TestRuntimeGODEBUG(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name        string
		program     string
		environment []string
		want        string
	}{
		{"compiled default", "testdata/runtimegodebug.go", nil, "true \"\"\ntrue \"\"\ntrue \"\"\ntrue \"tarinsecurepath=1\"\ntrue \"\"\ntrue \"\"\n"},
		{"environment override", "testdata/runtimegodebug.go", []string{"GODEBUG=tarinsecurepath=1"}, "true \"tarinsecurepath=1\"\ntrue \"\"\ntrue \"\"\ntrue \"tarinsecurepath=1\"\ntrue \"\"\ntrue \"\"\n"},
		{"setting compiled default", "testdata/runtimegodebugsetting.go", nil, "true\nfalse\ntrue\ntrue\n"},
		{"setting environment override", "testdata/runtimegodebugsetting.go", []string{"GODEBUG=tarinsecurepath=1"}, "false\nfalse\ntrue\ntrue\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			options := optionsFromTarget(*testTarget, sema)
			config, err := builder.NewConfig(&options)
			if err != nil {
				t.Fatal(err)
			}

			output := &bytes.Buffer{}
			_, err = buildAndRun(test.program, config, output, nil, test.environment, time.Minute, func(cmd *exec.Cmd, _ builder.BuildResult) error {
				return cmd.Run()
			})
			if err != nil {
				t.Fatal(err)
			}
			if output.String() != test.want {
				t.Fatalf("unexpected output:\n%s\nwant:\n%s", output.String(), test.want)
			}
		})
	}
}
