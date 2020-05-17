package wasm

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

var dockerHeadless = flag.Bool("docker-headless", false, "Launch headless Chrome via docker instead looking for it on local filesystem")
var noCleanup = flag.Bool("no-cleanup", false, "Skip cleanup operations (useful for manual viewing and debugging)")

func TestMain(m *testing.M) {
	flag.Parse()

	setup()
	ret := m.Run()
	cleanup()
	os.Exit(ret)
}

func setup() {

	// copy wasm_exec.js from targets to this dir
	b, err := ioutil.ReadFile("../../targets/wasm_exec.js")
	must(err)
	must(ioutil.WriteFile("wasm_exec.js", b, 0644))

	if *dockerHeadless {

		// build server for arch that works in docker
		args := []string{"go", "build", "-o", "server", "0setup_server.go"}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Env = append(os.Environ(), "GOARCH=amd64", "GOOS=linux")
		b, err := cmd.CombinedOutput()
		log.Printf("Command: %s; err=%v; full output:\n%s", strings.Join(args, " "), err, b)
		must(err)

		absDir, err := filepath.Abs(".")
		must(err)
		// launch headless chrome and server via docker if requested
		must(run("docker build -t tinygo_wasm_test -f 0dockerfile ."))
		run("docker stop tinygo_wasm_test") // ignore error

		must(runargs("docker", "run",
			"-d", "--rm",
			"--name=tinygo_wasm_test",
			"--mount", "type=bind,target=/wasm,source="+absDir,
			"-p", "9222:9222", "-p", "8826:8826",
			"tinygo_wasm_test:latest"))

	} else {

		killServer()

		// build server for local arch
		must(run("go build -o server 0setup_server.go"))

		// or otherwise just launch server
		serverCmd = exec.Command("./server")
		serverCmd.Stdout = os.Stdout
		serverCmd.Stderr = os.Stderr
		must(serverCmd.Start())

		must(ioutil.WriteFile("server.pid", []byte(fmt.Sprintf("%d\n", serverCmd.Process.Pid)), 0644))

	}

}

var serverCmd *exec.Cmd

func cleanup() {
	// log.Println("performing cleanup")

	if *noCleanup {
		return
	}
	if *dockerHeadless {
		run("docker stop tinygo_wasm_test") // ignore error
	} else {
		killServer()
	}
	os.Remove("wasm_exec.js")
	os.Remove("server")
	f, err := os.Open(".")
	must(err)
	defer f.Close()
	names, err := f.Readdirnames(-1)
	must(err)
	for _, name := range names {
		if filepath.Ext(name) == ".wasm" {
			os.Remove(name)
		}
	}
}

func run(cmdline string) error {
	args := strings.Fields(cmdline)
	return runargs(args...)
}

func runargs(args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	b, err := cmd.CombinedOutput()
	log.Printf("Command: %s; err=%v; full output:\n%s", strings.Join(args, " "), err, b)
	if err != nil {
		return err
	}
	return nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func chromectx(timeout time.Duration) (context.Context, context.CancelFunc) {

	var ctx context.Context

	if *dockerHeadless {

		// look up endpoint needed for chromedp (port 9222 is exposed by docker container)
		debugURL := func() string {
			resp, err := http.Get("http://localhost:9222/json/version")
			if err != nil {
				panic(err)
			}

			var result map[string]interface{}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				panic(err)
			}
			return result["webSocketDebuggerUrl"].(string)
		}()
		allocCtx, _ := chromedp.NewRemoteAllocator(context.Background(), debugURL)
		ctx, _ = chromedp.NewContext(allocCtx)

	} else {

		// looks for locally installed Chrome
		ctx, _ = chromedp.NewContext(context.Background())
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)

	return ctx, cancel
}

func killServer() {

	// log.Printf("killServer")

	defer os.Remove("server.pid")

	b, err := ioutil.ReadFile("server.pid")
	if err == nil {
		pid, _ := strconv.Atoi(strings.TrimSpace(string(b)))
		if pid > 0 {
			proc, err := os.FindProcess(pid)
			if err == nil {
				proc.Kill()
			}
		}
	}

}

func waitLog(logText string) chromedp.QueryAction {
	return waitInnerTextTrimEq("#log", logText)
}

// waitInnerTextTrimEq will wait for the innerText of the specified element to match a specific string after whitespace trimming.
func waitInnerTextTrimEq(sel, innerText string) chromedp.QueryAction {

	return chromedp.Query(sel, func(s *chromedp.Selector) {

		chromedp.WaitFunc(func(ctx context.Context, cur *cdp.Frame, ids ...cdp.NodeID) ([]*cdp.Node, error) {

			nodes := make([]*cdp.Node, len(ids))
			cur.RLock()
			for i, id := range ids {
				nodes[i] = cur.Nodes[id]
				if nodes[i] == nil {
					cur.RUnlock()
					// not yet ready
					return nil, nil
				}
			}
			cur.RUnlock()

			var ret string
			err := chromedp.EvaluateAsDevTools("document.querySelector('"+sel+"').innerText", &ret).Do(ctx)
			if err != nil {
				return nodes, err
			}
			if strings.TrimSpace(ret) != innerText {
				// log.Printf("found text: %s", ret)
				return nodes, errors.New("unexpected value: " + ret)
			}

			// log.Printf("NodeValue: %#v", nodes[0])

			// return nil, errors.New("not ready yet")
			return nodes, nil
		})(s)

	})

}
