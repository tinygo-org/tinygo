package wasm

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestFmt(t *testing.T) {

	t.Parallel()

	wasmTmpDir, server := startServer(t)

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
