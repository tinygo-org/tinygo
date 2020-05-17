package wasm

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestChan(t *testing.T) {

	err := run("tinygo build -o chan_pgm.wasm -target wasm chan_pgm.go")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := chromectx(5 * time.Second)
	defer cancel()

	err = chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8826/run?file=chan_pgm.wasm"),
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
