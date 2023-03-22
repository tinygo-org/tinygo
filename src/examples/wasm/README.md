# TinyGo WebAssembly examples

The examples here show two different ways of using WebAssembly with TinyGo:

1. Defining and exporting functions via the `//go:wasmimport <module>` directive. See
[the export folder](./export) for an example of this.  This exports the function to the Wasm
module (which has a default value of `env`). You can also use the Go directive
`//go:export <module>`.
1. Defining and executing a `func main()`. This is similar to how the Go
standard library implementation works. See [the main folder](./main) for an
example of this.

## Building

Build using the `tinygo` compiler:

```bash
$ tinygo build -o ./wasm.wasm -target wasm ./main/main.go
```

This creates a `wasm.wasm` file, which we can load in JavaScript and execute in
a browser.

Next, choose which example you want to use:
* [callback](callback): Defines and configures callbacks in Wasm.
* [export](export): Defines callbacks in Wasm, but configures them in JavaScript.
* [invoke](invoke): Invokes a function defined in JavaScript from Wasm.
* [main](main): Prints a message to the JavaScript console from Wasm.
* [slices](slices): Splits an Array defined in JavaScript from Wasm.

Let's say you chose [main](main), you'd build it like so:
```bash
$ make main
```

## Running

Start the local web server:

```bash
$ go run server.go
Serving ./html on http://localhost:8080
```

Use your web browser to visit http://localhost:8080.

* Tip: Open the browser development tools (e.g. Right-click, Inspect in
  FireFox) to see console output.

## How it works

Execution of the contents require a few JavaScript helper functions which are
called from WebAssembly.

We have defined these in [wasm_exec.js](../../../targets/wasm_exec.js).  It is
based on `$GOROOT/misc/wasm/wasm_exec.js` from the standard library, but is
slightly different. Ensure you are using the same version of `wasm_exec.js` as
the version of `tinygo` you are using to compile.

The general steps required to run the WebAssembly file in the browser includes
loading it into JavaScript with `WebAssembly.instantiateStreaming`, or
`WebAssembly.instantiate` in some browsers:

```js
const go = new Go(); // Defined in wasm_exec.js
const WASM_URL = 'wasm.wasm';

var wasm;

if ('instantiateStreaming' in WebAssembly) {
	WebAssembly.instantiateStreaming(fetch(WASM_URL), go.importObject).then(function (obj) {
		wasm = obj.instance;
		go.run(wasm);
	})
} else {
	fetch(WASM_URL).then(resp =>
		resp.arrayBuffer()
	).then(bytes =>
		WebAssembly.instantiate(bytes, go.importObject).then(function (obj) {
			wasm = obj.instance;
			go.run(wasm);
		})
	)
}
```

If you have used explicit exports, you can call them by invoking them under the
`wasm.exports` namespace. See the [`export`](./export/wasm.js) directory for an
example of this.

In addition to the JavaScript, it is important the wasm file is served with the
[`Content-Type`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type)
header set to `application/wasm`.  Without it, most browsers won't run it.

```go
package main

import (
	"log"
	"net/http"
	"strings"
)

const dir = "./html"

func main() {
	fs := http.FileServer(http.Dir(dir))
	log.Print("Serving " + dir + " on http://localhost:8080")
	http.ListenAndServe(":8080", http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Add("Cache-Control", "no-cache")
		if strings.HasSuffix(req.URL.Path, ".wasm") {
			resp.Header().Set("content-type", "application/wasm")
		}
		fs.ServeHTTP(resp, req)
	}))}
```

This simple server serves anything inside the `./html` directory on port
`8080`, setting any `*.wasm` files `Content-Type` header appropriately.

For development purposes (**only!**), it also sets the `Cache-Control` header
so your browser doesn't cache the files.  This is useful while developing, to
ensure your browser displays the newest wasm when you recompile.

In a production environment you **probably wouldn't** want to set the
`Cache-Control` header like this.  Caching is generally beneficial for end
users.

Further information on the `Cache-Control` header can be found here:

* https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
