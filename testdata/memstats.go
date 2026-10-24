package main

import "runtime"

func main() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	println("Lookups:", m.Lookups)
	println("StackInuse:", m.StackInuse)
	println("StackSys:", m.StackSys)
	println("MSpanInuse:", m.MSpanInuse)
	println("MSpanSys:", m.MSpanSys)
	println("MCacheInuse:", m.MCacheInuse)
	println("MCacheSys:", m.MCacheSys)
	println("BuckHashSys:", m.BuckHashSys)
	println("OtherSys:", m.OtherSys)
	println("NextGC:", m.NextGC)
	println("GCCPUFraction:", m.GCCPUFraction == 0)

	n, ok := runtime.ThreadCreateProfile(nil)
	println("ThreadCreateProfile:", n, ok)

	var r runtime.StackRecord
	r.Stack0[0] = 1
	r.Stack0[1] = 2
	println("Stack:", len(r.Stack()))
}
