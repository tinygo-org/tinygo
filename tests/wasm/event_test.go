package wasm

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestEvent(t *testing.T) {

	err := run("tinygo build -o event_pgm.wasm -target wasm event_pgm.go")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := chromectx(5 * time.Second)
	defer cancel()

	err = chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8826/run?file=event_pgm.wasm"),
		waitLog(`1
4`),
		chromedp.Click("#testbtn"),
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
