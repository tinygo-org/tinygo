package wasm

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestFmtprint(t *testing.T) {

	t.Parallel()

	wasmTmpDir, server := startServer(t)

	err := run("tinygo build -o " + wasmTmpDir + "/fmtprint.wasm -target wasm testdata/fmtprint.go")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := chromectx(5 * time.Second)
	defer cancel()

	var log1 string
	err = chromedp.Run(ctx,
		chromedp.Navigate(server.URL+"/run?file=fmtprint.wasm"),
		chromedp.Sleep(time.Second),
		chromedp.InnerHTML("#log", &log1),
		waitLog(`test from fmtprint 1
test from fmtprint 2
test from fmtprint 3
test from fmtprint 4`),
	)
	t.Logf("log1: %s", log1)
	if err != nil {
		t.Fatal(err)
	}

}
