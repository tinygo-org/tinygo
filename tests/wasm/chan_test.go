package wasm

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestChan(t *testing.T) {

	t.Parallel()

	err := run("tinygo build -o " + wasmTmpDir + "/chan.wasm -target wasm testdata/chan.go")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := chromectx(5 * time.Second)
	defer cancel()

	err = chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8826/run?file=chan.wasm"),
		waitLog(`1
2
4
3
true`),
	)
	if err != nil {
		t.Fatal(err)
	}

}
