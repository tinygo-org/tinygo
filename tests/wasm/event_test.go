package wasm

import (
	"testing"

	"github.com/chromedp/chromedp"
)

func TestEvent(t *testing.T) {

	wasmTmpDir, server := startServer(t)

	err := run(t, "tinygo build -o "+wasmTmpDir+"/event.wasm -target wasm testdata/event.go")
	if err != nil {
		t.Fatal(err)
	}

	ctx := chromectx(t)

	var log1, log2 string
	err = chromedp.Run(ctx,
		chromedp.Navigate(server.URL+"/run?file=event.wasm"),
		chromedp.WaitVisible("#log"),
		chromedp.InnerHTML("#log", &log1),
		waitLog(`1
4`),
		chromedp.Click("#testbtn"),
		chromedp.InnerHTML("#log", &log2),
		waitLog(`1
4
2
3
true`),
	)
	t.Logf("log1: %s", log1)
	t.Logf("log2: %s", log2)
	if err != nil {
		t.Fatal(err)
	}

}
