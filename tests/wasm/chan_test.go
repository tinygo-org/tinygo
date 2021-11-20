package wasm

import (
	"testing"

	"github.com/chromedp/chromedp"
)

func TestChan(t *testing.T) {

	wasmTmpDir, server := startServer(t)

	err := run(t, "tinygo build -o "+wasmTmpDir+"/chan.wasm -target wasm testdata/chan.go")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := chromectx()
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
