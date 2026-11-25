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

func TestAVRGCStackReturn(t *testing.T) {
	t.Parallel()

	options := optionsFromTarget("simavr", sema)
	emuCheck(t, options)
	config, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/main.go"
	err = os.WriteFile(path, []byte(`package main

import "runtime"

func main() {
	runtime.GC()
	println("GC returned")
}
`), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	output := &bytes.Buffer{}
	_, err = buildAndRun(path, config, output, nil, nil, 10*time.Second, func(cmd *exec.Cmd, result *builder.BuildResult) error {
		cmd.Stdout = output
		return cmd.Run()
	})
	if err != nil {
		t.Fatalf("GC program failed: %v\n%s", err, output.String())
	}
	if !strings.Contains(output.String(), "GC returned") {
		t.Fatalf("GC did not return:\n%s", output.String())
	}
}

// TestTimerStress runs timer and sleep stress programs. Timers fire while other
// goroutines stop, reset and create timers, sleep, and the GC runs. It covers
// the threads and tasks schedulers on the host, and the cores scheduler on
// riscv-qemu when qemu is installed. Programs that hold a timer callback in
// flight need a second core to keep running, so only the cores scheduler runs
// them.
func TestTimerStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timer stress tests in short mode")
	}
	programs := []string{
		"timer_queue_growth.go",
		"timer_concurrent_growth.go",
		"timer_stop_reset_stress.go",
		"timer_ticker_gc.go",
		"timer_growth_under_gc.go",
		"sleepers.go",
	}
	coresPrograms := []string{
		"timer_stop_reset_race.go",
		"timer_reserve_alloc.go",
		"timer_reserve_alloc_multicore.go",
	}
	type config struct {
		name      string
		target    string
		scheduler string
	}
	var configs []config
	switch runtime.GOOS {
	case "darwin", "linux":
		configs = append(configs, config{"threads", "", "threads"}, config{"tasks", "", "tasks"})
	}
	if _, err := exec.LookPath("qemu-system-riscv32"); err == nil {
		configs = append(configs, config{"riscv-qemu", "riscv-qemu", ""})
	}
	if len(configs) == 0 {
		t.Skip("no supported scheduler configuration on this host")
	}
	for _, c := range configs {
		cfgPrograms := programs
		if c.target == "riscv-qemu" {
			cfgPrograms = append(cfgPrograms[:len(cfgPrograms):len(cfgPrograms)], coresPrograms...)
		}
		for _, program := range cfgPrograms {
			t.Run(c.name+"/"+program, func(t *testing.T) {
				options := optionsFromTarget(c.target, sema)
				options.Scheduler = c.scheduler
				if c.target == "riscv-qemu" && program == "timer_stop_reset_stress.go" {
					// Short length still takes over 2 minutes on assert builds in CI.
					options.GlobalValues = map[string]map[string]string{"main": {"size": "small"}}
				}
				runTest(program, options, t, nil, nil)
			})
		}
	}
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
			_, err = buildAndRun("testdata/goexit.go", config, output, []string{tc.name}, nil, time.Minute, func(cmd *exec.Cmd, result *builder.BuildResult) error {
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
	_, err = buildAndRun("testdata/runtimefatal.go", config, output, nil, nil, time.Minute, func(cmd *exec.Cmd, result *builder.BuildResult) error {
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
	_, err = buildAndRun(TESTDATA+"/stdin.go", config, output, nil, nil, time.Minute, func(cmd *exec.Cmd, result *builder.BuildResult) error {
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
			_, err = buildAndRun("testdata/runtimegoroot.go", config, output, nil, test.environment, time.Minute, func(cmd *exec.Cmd, _ *builder.BuildResult) error {
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
			_, err = buildAndRun(test.program, config, output, nil, test.environment, time.Minute, func(cmd *exec.Cmd, _ *builder.BuildResult) error {
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

func TestTimerChannelGC(t *testing.T) {
	t.Parallel()
	switch runtime.GOOS {
	case "darwin", "linux":
	default:
		t.Skipf("host GOOS %s does not use the Boehm GC", runtime.GOOS)
	}
	options := optionsFromTarget("", sema)
	// Enumeration must not race runtime allocations.
	// BDWGC is built without GC_THREADS in builder/bdwgc.go.
	options.Scheduler = "tasks"
	runTest("timer_gc.go", options, t, nil, nil)
}
