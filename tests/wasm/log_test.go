package wasm

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestLog(t *testing.T) {

	t.Parallel()

	wasmTmpDir, server := startServer(t)

	err := run("tinygo build -o " + wasmTmpDir + "/log.wasm -target wasm testdata/log.go")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := chromectx(5 * time.Second)
	defer cancel()

	var log1 string
	err = chromedp.Run(ctx,
		chromedp.Navigate(server.URL+"/run?file=log.wasm"),
		chromedp.Sleep(time.Second),
		chromedp.InnerHTML("#log", &log1),
		waitLogRe(`^..../../.. ..:..:.. log 1
..../../.. ..:..:.. log 2
..../../.. ..:..:.. log 3
println 4
fmt.Println 5
..../../.. ..:..:.. log 6
in func 1
..../../.. ..:..:.. in func 2
$`),
	)
	t.Logf("log1: %s", log1)
	if err != nil {
		t.Fatal(err)
	}

}
