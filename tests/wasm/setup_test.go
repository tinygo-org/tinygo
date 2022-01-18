package wasm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func run(t *testing.T, cmdline string) error {
	args := strings.Fields(cmdline)
	return runargs(t, args...)
}

func runargs(t *testing.T, args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	b, err := cmd.CombinedOutput()
	t.Logf("Command: %s; err=%v; full output:\n%s", strings.Join(args, " "), err, b)
	if err != nil {
		return err
	}
	return nil
}

func chromectx(t *testing.T) context.Context {
	// looks for locally installed Chrome
	ctx, ccancel := chromedp.NewContext(context.Background(), chromedp.WithErrorf(t.Errorf), chromedp.WithDebugf(t.Logf), chromedp.WithLogf(t.Logf))
	t.Cleanup(ccancel)

	// Wait for browser to be ready.
	err := chromedp.Run(ctx)
	if err != nil {
		t.Fatalf("failed to start browser: %s", err.Error())
	}

	ctx, tcancel := context.WithTimeout(ctx, 30*time.Second)
	t.Cleanup(tcancel)

	return ctx
}

func startServer(t *testing.T) (string, *httptest.Server) {
	tmpDir := t.TempDir()

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

	server := httptest.NewServer(h)
	t.Logf("Started server at %q for dir: %s", server.URL, tmpDir)
	t.Cleanup(server.Close)

	return tmpDir, server
}

// waitLog blocks until the log output equals the text provided (ignoring whitespace before and after)
func waitLog(logText string) chromedp.QueryAction {
	return waitInnerTextTrimEq("#log", strings.TrimSpace(logText))
}

// waitLogRe blocks until the log output matches this regular expression
func waitLogRe(restr string) chromedp.QueryAction {
	return waitInnerTextMatch("#log", regexp.MustCompile(restr))
}

// waitInnerTextTrimEq will wait for the innerText of the specified element to match a specific text pattern (ignoring whitespace before and after)
func waitInnerTextTrimEq(sel string, innerText string) chromedp.QueryAction {
	return waitInnerTextMatch(sel, regexp.MustCompile(`^\s*`+regexp.QuoteMeta(innerText)+`\s*$`))
}

// waitInnerTextMatch will wait for the innerText of the specified element to match a specific regexp pattern
func waitInnerTextMatch(sel string, re *regexp.Regexp) chromedp.QueryAction {

	return chromedp.Query(sel, func(s *chromedp.Selector) {

		chromedp.WaitFunc(func(ctx context.Context, cur *cdp.Frame, execCtx runtime.ExecutionContextID, ids ...cdp.NodeID) ([]*cdp.Node, error) {

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
			if !re.MatchString(ret) {
				// log.Printf("found text: %s", ret)
				return nodes, errors.New("unexpected value: " + ret)
			}

			// log.Printf("NodeValue: %#v", nodes[0])

			// return nil, errors.New("not ready yet")
			return nodes, nil
		})(s)

	})

}
