// +build go1.14

package wasm

// NOTE: this should work in go1.13 but panics with:
// panic: syscall/js: call of Value.Get on string
// which is coming from here: https://github.com/golang/go/blob/release-branch.go1.13/src/syscall/js/js.go#L252
// But I'm not sure how import "fmt" results in this.
// To reproduce, install Go 1.13.x and change the build tag above
// to go1.13 and run this test.

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestFmt(t *testing.T) {

	t.Parallel()

	wasmTmpDir, server, cleanup := startServer(t)
	defer cleanup()

	err := run("tinygo build -o " + wasmTmpDir + "/fmt.wasm -target wasm testdata/fmt.go")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := chromectx(5 * time.Second)
	defer cancel()

	var log1 string
	err = chromedp.Run(ctx,
		chromedp.Navigate(server.URL+"/run?file=fmt.wasm"),
		chromedp.Sleep(time.Second),
		chromedp.InnerHTML("#log", &log1),
		waitLog(`did not panic`),
	)
	t.Logf("log1: %s", log1)
	if err != nil {
		t.Fatal(err)
	}

}
