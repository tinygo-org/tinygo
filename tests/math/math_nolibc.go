//go:build nolibc

package math_test

//go:export exp
func exp(f float64) float64 {
	return mathExp(f)
}

//go:linkname mathExp math.exp
func mathExp(float64) float64

//go:export exp2
func exp2(f float64) float64 {
	return mathExp(f)
}

//go:linkname mathExp2 math.exp2
func mathExp2(float64) float64

//go:export log
func log(f float64) float64 {
	return mathLog(f)
}

//go:linkname mathLog math.log
func mathLog(float64) float64
