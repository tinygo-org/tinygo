package wasm

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestChan(t *testing.T) {

	t.Parallel()

	wasmTmpDir, server := startServer(t)

	err := run("tinygo build -o " + wasmTmpDir + "/chan.wasm -target wasm testdata/chan.go")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := chromectx(5 * time.Second)
	defer cancel()

	err = chromedp.Run(ctx,
		chromedp.Navigate(server.URL+"/run?file=chan.wasm"),
		waitLog(`1
4
2
3
true`),
	)
	if err != nil {
		t.Fatal(err)
	}
}
