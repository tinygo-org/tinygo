// qbs runs quick, broad, shallow tests on a connected board. It flashes a
// program for each class of tests, collects the results over serial and
// checks them.
//
//	go run ./tools/qbs -target pico
//	go run ./tools/qbs -target arduino-uno-r4-wifi timing gc
//
// Classes:
//
//	timing   sleep accuracy and the device clock against the host clock
//	compute  results of the same computations on the device and on the host
//	gc       allocation, collection and fragmentation checks
//	serial   echo, RX overrun, line and byte transfer on the default serial port
//
// With -emulator the programs run in the emulator of the target through
// tinygo run, for example -target wasip1, and no board is needed.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"

	"github.com/tinygo-org/tinygo/tools/qbs/cases"
)

type tick struct {
	device float64 // seconds on the device clock
	host   time.Time
}

type testResult struct {
	name    string
	status  string
	detail  string
	elapsed time.Duration
	vals    [][2]string
	logs    []string
	ticks   []tick
	stray   []string
	raw     map[string][]byte
	// hostFail is set by a driver that found a problem while it ran.
	hostFail string
}

// driver runs on the host when a test begins. It may send data to the device
// and read from it.
type driver func(l *link, t *testResult)

type classResult struct {
	class string
	tests []*testResult
	done  bool
	err   string
	raw   []byte
}

type link struct {
	port conn
	buf  []byte
	log  []byte
}

func (l *link) fill() {
	var tmp [512]byte
	n, _ := l.port.Read(tmp[:])
	if n > 0 {
		l.buf = append(l.buf, tmp[:n]...)
		l.log = append(l.log, tmp[:n]...)
	}
}

// readLine returns the next line and the time it was complete.
func (l *link) readLine(timeout time.Duration) (string, time.Time, bool) {
	deadline := time.Now().Add(timeout)
	for {
		for i, c := range l.buf {
			if c == '\n' {
				line := strings.TrimRight(string(l.buf[:i]), "\r")
				l.buf = l.buf[i+1:]
				return cleanConsole(line), time.Now(), true
			}
		}
		if time.Now().After(deadline) {
			return "", time.Time{}, false
		}
		l.fill()
	}
}

// readExact returns n bytes, or what arrived before the timeout.
func (l *link) readExact(n int, timeout time.Duration) []byte {
	deadline := time.Now().Add(timeout)
	for len(l.buf) < n && time.Now().Before(deadline) {
		l.fill()
	}
	if len(l.buf) < n {
		n = len(l.buf)
	}
	out := append([]byte(nil), l.buf[:n]...)
	l.buf = l.buf[n:]
	return out
}

var ansiEscape = regexp.MustCompile("\x1b\\[[0-9;]*m")

// cleanConsole removes what the simavr console adds to a line. It colours the
// line and shows the end of line characters as two dots.
func cleanConsole(line string) string {
	if !strings.Contains(line, "\x1b[") {
		return line
	}
	line = ansiEscape.ReplaceAllString(line, "")
	return strings.TrimSuffix(line, "..")
}

func openPort(name string) (*link, error) {
	deadline := time.Now().Add(15 * time.Second)
	var err error
	for time.Now().Before(deadline) {
		var p serial.Port
		p, err = serial.Open(name, &serial.Mode{BaudRate: 115200})
		if err == nil {
			p.SetReadTimeout(50 * time.Millisecond)
			return &link{port: p}, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil, fmt.Errorf("cannot open %s: %v", name, err)
}

func findPort() string {
	for _, pattern := range []string{"/dev/cu.usbmodem*", "/dev/ttyACM*", "/dev/cu.usbserial*", "/dev/ttyUSB*"} {
		if ports, _ := filepath.Glob(pattern); len(ports) > 0 {
			return ports[0]
		}
	}
	return ""
}

type options struct {
	tinygo string
	target string
	// port is the serial port the results are read from.
	port string
	// flashPort is passed to tinygo flash. It is empty unless the user chose
	// a port, so that tinygo finds the board itself.
	flashPort   string
	pathPrepend string
	// noInput is set for emulated targets that cannot receive input.
	noInput bool
	// pty runs the emulator under a pseudo terminal, so that emulators that
	// buffer their output when it is a pipe print each line at once.
	pty bool
	// scheduler overrides the scheduler the programs are built with.
	scheduler string
	// gc overrides the garbage collector the programs are built with.
	gc string
}

// flash returns skip with a reason when the program does not fit on the target.
// It waits up to 90 seconds for the drive of a UF2 bootloader to be mounted.
func flash(o options, dir string) (skip string, err error) {
	args := []string{"flash", "-target", o.target, "-timeout", "90s"}
	if o.flashPort != "" {
		args = append(args, "-port", o.flashPort)
	}
	cmd := exec.Command(o.tinygo, append(args, dir)...)
	cmd.Dir = dir
	if o.pathPrepend != "" {
		cmd.Env = append(os.Environ(), "PATH="+o.pathPrepend+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
	out, runErr := cmd.CombinedOutput()
	if runErr == nil {
		return "", nil
	}
	text := string(out)
	if reason := fitFailure(text); reason != "" {
		return reason, nil
	}
	if len(text) > 1500 {
		text = text[len(text)-1500:]
	}
	return "", fmt.Errorf("flash failed: %v\n%s", runErr, text)
}

// classTimeout limits how long a class may take in all. A program can keep
// printing, for example garbage after it runs out of memory, so the silence
// timeout alone does not end it.
var classTimeout = 10 * time.Minute

// readyTimeout is how long to wait for the first banner of a program. An
// emulated program is compiled first, which takes longer.
var readyTimeout = 25 * time.Second

// collect runs the handshake and reads results until the device reports the end.
func collect(l *link, class string, drivers map[string]driver, noInput bool) *classResult {
	r := &classResult{class: class}
	defer func() { r.raw = l.log }()

	// A program that cannot receive input starts without the host, so wait
	// for its first line and send nothing.
	banner := "@@READY qbs " + class
	if noInput {
		banner = "@@GO " + class
	}
	ready := false
	deadline := time.Now().Add(readyTimeout)
	for time.Now().Before(deadline) {
		line, _, ok := l.readLine(time.Second)
		if ok && strings.HasPrefix(line, banner) {
			ready = true
			break
		}
		// An emulated program that already ended will not print a banner.
		if e, isPipe := l.port.(interface{ exited() bool }); isPipe && e.exited() {
			break
		}
	}
	if !ready {
		r.err = "no " + banner + " line, the program printed nothing"
		return r
	}
	if !noInput {
		l.port.Write([]byte("G"))
	}

	// The first @@GO is the banner when the program does not wait for the host.
	started := noInput
	var cur *testResult
	classDeadline := time.Now().Add(classTimeout)
	for {
		if time.Now().After(classDeadline) {
			where := "before the first test"
			if cur != nil {
				where = "in test " + cur.name
			}
			r.err = "the class did not finish within " + classTimeout.String() + ", it was " + where
			return r
		}
		line, at, ok := l.readLine(60 * time.Second)
		if !ok {
			where := "before the first test"
			if cur != nil {
				where = "in test " + cur.name
			}
			r.err = "no output for 60s " + where
			return r
		}
		f := strings.Fields(line)
		if len(f) == 0 || !strings.HasPrefix(f[0], "@@") {
			if cur != nil {
				cur.stray = append(cur.stray, line)
			}
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, f[0]))
		switch f[0] {
		case "@@GO":
			// A second one means the program restarted, for example after a crash.
			if started {
				where := "before the first test"
				if cur != nil {
					where = "in test " + cur.name
				}
				r.err = "the program restarted " + where
				return r
			}
			started = true
		case "@@BEGIN":
			cur = &testResult{name: rest, raw: map[string][]byte{}}
			r.tests = append(r.tests, cur)
			if d := drivers[cur.name]; d != nil {
				d(l, cur)
			}
		case "@@LOG":
			if cur != nil {
				cur.logs = append(cur.logs, rest)
			}
		case "@@VAL":
			if cur != nil && len(f) >= 2 {
				cur.vals = append(cur.vals, [2]string{f[1], strings.TrimSpace(strings.TrimPrefix(rest, f[1]))})
			}
		case "@@RAW":
			if cur != nil && len(f) == 3 {
				n, err := strconv.Atoi(f[2])
				if err == nil {
					cur.raw[f[1]] = l.readExact(n, 10*time.Second)
				}
			}
		case "@@TICK":
			if cur != nil && len(f) == 3 {
				if us, err := strconv.ParseFloat(f[2], 64); err == nil {
					cur.ticks = append(cur.ticks, tick{device: us / 1e6, host: at})
				}
			}
		case "@@RESULT":
			if cur != nil && len(f) >= 4 && f[1] == cur.name {
				cur.status = f[2]
				us, _ := strconv.ParseInt(f[3], 10, 64)
				cur.elapsed = time.Duration(us) * time.Microsecond
				cur.detail = strings.TrimSpace(strings.Join(f[4:], " "))
			}
		case "@@END":
			r.done = true
			return r
		}
	}
}

func (t *testResult) val(key string) (string, bool) {
	for _, kv := range t.vals {
		if kv[0] == key {
			return kv[1], true
		}
	}
	return "", false
}

func (t *testResult) fail(format string, args ...interface{}) {
	t.status = "FAIL"
	t.detail = fmt.Sprintf(format, args...)
}

// emulated is set with -emulator. An emulator does not keep wall clock time.
var emulated bool

// evalTiming compares the device clock with the host clock.
func evalTiming(r *classResult) {
	for _, t := range r.tests {
		if t.name != "clock-rate" || t.status != "PASS" {
			continue
		}
		if len(t.ticks) < 20 {
			t.fail("only %d of 30 ticks arrived", len(t.ticks))
			continue
		}
		var mx, my float64
		t0 := t.ticks[0].host
		for _, k := range t.ticks {
			mx += k.host.Sub(t0).Seconds()
			my += k.device
		}
		n := float64(len(t.ticks))
		mx /= n
		my /= n
		var sxy, sxx float64
		for _, k := range t.ticks {
			x := k.host.Sub(t0).Seconds() - mx
			sxy += x * (k.device - my)
			sxx += x * x
		}
		slope := sxy / sxx
		if emulated && slope > 10 {
			t.status = "SKIP"
			t.detail = fmt.Sprintf("the device clock runs %.0f times faster than the host clock, it is probably simulated", slope)
			continue
		}
		t.detail = fmt.Sprintf("device clock runs at %.4f of the host clock", slope)
		if math.Abs(slope-1) > 0.02 {
			if emulated {
				t.status = "SKIP"
				t.detail = fmt.Sprintf("the emulated clock runs at %.4f of the host clock", slope)
				continue
			}
			t.fail("device clock runs at %.4f of the host clock", slope)
		}
	}
}

// compareCase compares a device result with the host result.
func compareCase(c cases.Case, got, want string) (bool, string) {
	if got == want {
		return true, ""
	}
	if c.ULP == 0 {
		return false, "device " + shorten(got) + " host " + shorten(want)
	}
	g, w := strings.Fields(got), strings.Fields(want)
	if len(g) != len(w) {
		return false, fmt.Sprintf("device returned %d values, host %d", len(g), len(w))
	}
	var worst uint64
	for i := range g {
		a, errA := strconv.ParseUint(g[i], 16, 64)
		b, errB := strconv.ParseUint(w[i], 16, 64)
		if errA != nil || errB != nil {
			return false, "device " + shorten(got) + " host " + shorten(want)
		}
		d := a - b
		if b > a {
			d = b - a
		}
		if d > worst {
			worst = d
		}
		if d > c.ULP {
			return false, fmt.Sprintf("value %d differs by %d ULP (limit %d)", i, d, c.ULP)
		}
	}
	return true, fmt.Sprintf("within %d ULP", worst)
}

func shorten(s string) string {
	if len(s) > 60 {
		return s[:57] + "..."
	}
	return s
}

// evalCompute runs the same cases on the host and compares.
func evalCompute(r *classResult) {
	byName := map[string]cases.Case{}
	for _, c := range cases.All() {
		byName[c.Name] = c
	}
	for _, t := range r.tests {
		c, ok := byName[t.name]
		if !ok || t.status != "PASS" {
			continue
		}
		got, ok := t.val("result")
		if !ok {
			t.fail("device sent no result")
			continue
		}
		match, why := compareCase(c, got, c.Fn())
		if !match {
			t.fail("%s", why)
		} else {
			t.detail = why
		}
	}
}

var classes = []struct {
	name       string
	drivers    map[string]driver
	eval       func(*classResult)
	needsInput bool
}{
	{"timing", nil, evalTiming, false},
	{"compute", nil, evalCompute, false},
	{"gc", nil, nil, false},
	{"serial", serialDrivers, evalSerial, true},
}

func main() {
	target := flag.String("target", "", "tinygo target (required)")
	tinygo := flag.String("tinygo", "tinygo", "tinygo binary to use")
	port := flag.String("port", "", "serial port (default: first USB serial port found)")
	pathPrepend := flag.String("path-prepend", "", "directory to put first on PATH when flashing")
	verbose := flag.Bool("v", false, "print the values and messages each test sent")
	dump := flag.Bool("dump", false, "print raw serial data when a class fails to finish")
	emulate := flag.Bool("emulator", false, "run the programs in the emulator of the target instead of flashing a board")
	usePTY := flag.Bool("pty", false, "with -emulator, run the emulator under a pseudo terminal, which simavr needs to print lines as they happen")
	scheduler := flag.String("scheduler", "", "with -emulator, build the programs with this scheduler, for example none")
	flag.DurationVar(&classTimeout, "class-timeout", classTimeout, "longest time one class may take")
	gcFlag := flag.String("gc", "", "with -emulator, build the programs with this garbage collector, for example precise")
	noInput := flag.Bool("noinput", false, "with -emulator, the target cannot receive input, so run only the classes that need none")
	flag.Parse()

	if *target == "" {
		fmt.Fprintln(os.Stderr, "missing -target")
		os.Exit(2)
	}
	flashPort := *port
	emulated = *emulate
	if *port == "" && !*emulate {
		*port = findPort()
	}
	if *port == "" && !*emulate {
		fmt.Fprintln(os.Stderr, "no serial port found, use -port")
		os.Exit(2)
	}
	want := map[string]bool{}
	for _, name := range flag.Args() {
		want[name] = true
	}
	for name := range want {
		known := false
		for _, c := range classes {
			known = known || c.name == name
		}
		if !known {
			fmt.Fprintf(os.Stderr, "unknown class %q\n", name)
			os.Exit(2)
		}
	}

	_, file, _, _ := runtime.Caller(0)
	here := filepath.Dir(file)
	o := options{tinygo: *tinygo, target: *target, port: *port, flashPort: flashPort, pathPrepend: *pathPrepend, noInput: *noInput, pty: *usePTY, scheduler: *scheduler, gc: *gcFlag}

	fmt.Printf("port %s, target %s\n", *port, *target)
	counts := map[string]int{}
	report := func(status, class, name, detail string) {
		counts[status]++
		fmt.Printf("%-5s %-8s %-16s %s\n", status, class, name, detail)
	}

	for _, c := range classes {
		if len(want) > 0 && !want[c.name] {
			continue
		}
		fmt.Printf("\n== %s\n", c.name)
		if *noInput && c.needsInput {
			report("SKIP", c.name, "(class)", "the target cannot receive input")
			continue
		}
		var l *link
		if *emulate {
			readyTimeout = 3 * time.Minute
			var err error
			l, err = startEmulator(o, filepath.Join(here, c.name))
			if err != nil {
				if reason := emulatorFailure([]byte(err.Error())); reason != "" {
					report("SKIP", c.name, "(class)", reason)
				} else {
					report("FAIL", c.name, "(emulator)", err.Error())
				}
				continue
			}
		} else {
			flashStart := time.Now()
			skip, err := flash(o, filepath.Join(here, c.name))
			if took := time.Since(flashStart); took > 15*time.Second {
				fmt.Printf("note: flashing took %s\n", took.Round(time.Second))
			}
			if err != nil {
				report("FAIL", c.name, "(flash)", err.Error())
				continue
			}
			if skip != "" {
				report("SKIP", c.name, "(class)", skip)
				continue
			}
			time.Sleep(time.Second)
			l, err = openPort(*port)
			if err != nil {
				report("FAIL", c.name, "(port)", err.Error())
				continue
			}
		}
		r := collect(l, c.name, c.drivers, *noInput)
		l.port.Close()
		if *emulate {
			if reason := emulatorFailure(r.raw); reason != "" {
				report("SKIP", c.name, "(class)", reason)
				continue
			}
		}

		if c.eval != nil {
			c.eval(r)
		}
		for _, t := range r.tests {
			status := t.status
			detail := t.detail
			if status == "" {
				status = "FAIL"
				detail = "no result, the device stopped in this test"
				if len(t.stray) > 0 {
					detail += ": " + shorten(strings.Join(t.stray, " | "))
				}
			} else if status == "PASS" && detail == "" {
				detail = t.elapsed.Round(time.Millisecond).String()
			}
			report(status, c.name, t.name, detail)
			if *verbose {
				for _, kv := range t.vals {
					fmt.Printf("        %s = %s\n", kv[0], shorten(kv[1]))
				}
				for _, m := range t.logs {
					fmt.Printf("        log: %s\n", m)
				}
			}
		}
		if r.err != "" {
			report("FAIL", c.name, "(class)", r.err)
			if *dump {
				raw := r.raw
				if len(raw) > 400 {
					raw = raw[len(raw)-400:]
				}
				fmt.Printf("raw serial data: %q\n", raw)
			}
		}
	}

	fmt.Printf("\n%d pass, %d fail, %d skip\n", counts["PASS"], counts["FAIL"], counts["SKIP"])
	if counts["FAIL"] > 0 {
		os.Exit(1)
	}
}
