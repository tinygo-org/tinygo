// +build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
)

// This is a standalone static file server which is used to serve pages for browser testing.
// It is a separate program so we can run it inside the same docker container as headless chrome,
// in order to avoid port mapping issues.

var addr = flag.String("addr", ":8826", "Host:port to listen on")
var dir = flag.String("dir", ".", "Directory to serve static files from")

func main() {

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		panic(err)
	}
	fsh := http.FileServer(http.Dir(absDir))
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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

	log.Printf("Starting server at %q for dir: %s", *addr, absDir)
	log.Fatal(http.ListenAndServe(*addr, h))
}
