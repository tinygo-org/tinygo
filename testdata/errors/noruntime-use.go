package main

//go:noruntime
func calc(n int) int { return n * 7 }

func main() {
	println(calc(12))
}

// ERROR: noruntime-use.go:4: //go:noruntime function main.calc remains reachable at runtime
// ERROR: noruntime-use.go:7:14: //go:noruntime function main.calc is used at runtime by main.main
