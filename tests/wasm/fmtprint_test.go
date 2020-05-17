// +build go1.14

package wasm

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestFmtprint(t *testing.T) {

	err := run("tinygo build -o fmtprint_pgm.wasm -target wasm fmtprint_pgm.go")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := chromectx(5 * time.Second)
	defer cancel()

	var log1 string
	err = chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8826/run?file=fmtprint_pgm.wasm"),
		chromedp.Sleep(time.Second),
		chromedp.InnerHTML("#log", &log1),
		waitLog(`test from fmtprint_pgm 1
test from fmtprint_pgm 2
test from fmtprint_pgm 3
test from fmtprint_pgm 4`),
	)
	t.Logf("log1: %s", log1)
	if err != nil {
		t.Fatal(err)
	}

}
