// +build go1.12

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

	var log1, log2 string
	err = chromedp.Run(ctx,
		chromedp.Navigate("http://localhost:8826/run?file=event_pgm.wasm"),
		chromedp.WaitVisible("#log"),
		chromedp.Sleep(time.Second),
		chromedp.InnerHTML("#log", &log1),
		waitLog(`1
4`),
		chromedp.Click("#testbtn"),
		chromedp.Sleep(time.Second),
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
