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

	err = chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8826/run?file=fmtprint_pgm.wasm"),
		waitLog(`test from fmtprint_pgm 1
test from fmtprint_pgm 2
test from fmtprint_pgm 3
test from fmtprint_pgm 4`),
	)
	if err != nil {
		t.Fatal(err)
	}

}
