package wasm

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

var addr = flag.String("addr", ":8826", "Host:port to listen on for wasm test server")

var wasmTmpDir string // set in TestMain to a temp directory for build output

func TestMain(m *testing.M) {
	flag.Parse()

	var err error
	wasmTmpDir, err = ioutil.TempDir("", "wasm_test")
	if err != nil {
		log.Fatalf("unable to create temp dir: %v", err)
	}

	startServer(wasmTmpDir)

	var ret int
	func() {
		defer os.RemoveAll(wasmTmpDir) // cleanup even on panic
		ret = m.Run()
	}()

	os.Exit(ret)
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

func chromectx(timeout time.Duration) (context.Context, context.CancelFunc) {

	var ctx context.Context

	// looks for locally installed Chrome
	ctx, _ = chromedp.NewContext(context.Background())

	ctx, cancel := context.WithTimeout(ctx, timeout)

	return ctx, cancel
}

func startServer(tmpDir string) {

	fsh := http.FileServer(http.Dir(tmpDir))
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == "/wasm_exec.js" {
			http.ServeFile(w, r, "../../targets/wasm_exec.js")
			return
		}

		if r.URL.Path == "/run" {
			fmt.Fprintf(w, `<!doctype html>
<html>
<head>
<title>Test</title>
<meta charset="utf-8"/>
</head>
<body>
<div id="main"></div>
<pre id="log"></pre>
<script>
window.wasmLogOutput  = [];
(function() {
	var logdiv = document.getElementById('log');
	var cl = console.log;
	console.log = function() {
		var a = [];
		for (var i = 0; i < arguments.length; i++) {
			a.push(arguments[i]);
		}
		var line = a.join(' ') + "\n";
		window.wasmLogOutput.push(line);
		var ret = cl.apply(console, arguments)
		var el = document.createElement('span');
		el.innerText = line;
		logdiv.appendChild(el);
		return ret
	}
})()
</script>
<script src="/wasm_exec.js"></script>
<script>
var wasmSupported = (typeof WebAssembly === "object");
if (wasmSupported) {
	var mainWasmReq = fetch("/%s").then(function(res) {
		if (res.ok) {
			const go = new Go();
			WebAssembly.instantiateStreaming(res, go.importObject).then((result) => {
				go.run(result.instance);
			});		
		} else {
			res.text().then(function(txt) {
				var el = document.getElementById("main");
				el.style = 'font-family: monospace; background: black; color: red; padding: 10px';
				el.innerText = txt;
			})
		}
	})
} else {
	document.getElementById("main").innerHTML = 'This application requires WebAssembly support.  Please upgrade your browser.';
}
</script>
</body>
</html>`, r.FormValue("file"))
			return
		}

		fsh.ServeHTTP(w, r)
	})

	log.Printf("Starting server at %q for dir: %s", *addr, tmpDir)
	go func() {
		log.Fatal(http.ListenAndServe(*addr, h))
	}()

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
