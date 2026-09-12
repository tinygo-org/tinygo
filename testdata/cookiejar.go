package main

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

func main() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	u, err := url.Parse("https://example.com/account")
	if err != nil {
		panic(err)
	}
	jar.SetCookies(u, []*http.Cookie{{Name: "session", Value: "value", Secure: true, Path: "/"}})
	if cookies := jar.Cookies(u); len(cookies) != 1 || cookies[0].Value != "value" {
		panic("cookie not stored")
	}
	u.Scheme = "http"
	if len(jar.Cookies(u)) != 0 {
		panic("secure cookie sent over HTTP")
	}
	println("cookie stored; secure cookie refused over HTTP")
}
