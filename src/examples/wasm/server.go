package main

import (
	"log"
	"net/http"
	"strings"
)

const dir = "./html"

func main() {
	fs := http.FileServer(http.Dir(dir))
	log.Print("Serving " + dir + " on http://localhost:8080")
	http.ListenAndServe(":8080", http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Cache-Control", "max-age=0")
		resp.Header().Add("Cache-Control", "no-store")
		resp.Header().Add("Cache-Control", "no-cache")
		resp.Header().Add("Cache-Control", "must-revalidate")
		resp.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
		if strings.HasSuffix(req.URL.Path, ".wasm") {
			resp.Header().Set("content-type", "application/wasm")
		}

		fs.ServeHTTP(resp, req)
	}))
}
