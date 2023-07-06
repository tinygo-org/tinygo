package main

func min1(a int) int {
	return min(a)
}

func min2(a, b int) int {
	return min(a, b)
}

func min3(a, b, c int) int {
	return min(a, b, c)
}

func min4(a, b, c, d int) int {
	return min(a, b, c, d)
}

func minUint8(a, b uint8) uint8 {
	return min(a, b)
}

func minUnsigned(a, b uint) uint {
	return min(a, b)
}

func minFloat32(a, b float32) float32 {
	return min(a, b)
}

func minFloat64(a, b float64) float64 {
	return min(a, b)
}

func minString(a, b string) string {
	return min(a, b)
}

func maxInt(a, b int) int {
	return max(a, b)
}

func maxUint(a, b uint) uint {
	return max(a, b)
}

func maxFloat32(a, b float32) float32 {
	return max(a, b)
}

func maxString(a, b string) string {
	return max(a, b)
}
