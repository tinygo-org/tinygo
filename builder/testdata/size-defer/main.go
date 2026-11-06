package main

var locked bool
var checksum uint32

//go:noinline
func acquire() { locked = true }

//go:noinline
func release() { locked = false }

//go:noinline
func rotate(x uint32) uint32 { return x<<5 | x>>27 }

//go:noinline
func mix(a, b uint32) uint32 { return (a ^ b) * 16777619 }

//go:noinline
func clamp(x, limit uint32) uint32 {
	if x > limit {
		return limit
	}
	return x
}

//go:noinline
func header(n uint32) uint32 {
	acquire()
	defer release()
	h := rotate(n)
	h = mix(h, n+1)
	return clamp(h, 1500)
}

//go:noinline
func payload(n uint32) uint32 {
	acquire()
	defer release()
	sum := mix(0, rotate(n))
	sum = mix(sum, clamp(n, 64))
	return mix(sum, rotate(sum))
}

func main() {
	for i := uint32(0); i < 4; i++ {
		checksum = mix(checksum, header(i))
		checksum = mix(checksum, payload(i))
	}
	println(checksum)
}
