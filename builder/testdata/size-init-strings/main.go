package main

import "strings"

var banner = strings.Repeat("=", 16) + "\n" + strings.Join([]string{"sensor", "v1.0"}, " ")

var escaper = strings.NewReplacer("\r", `\r`, "\n", `\n`)

func main() {
	println(banner)
	println(escaper.Replace("a\r\nb"))
}
