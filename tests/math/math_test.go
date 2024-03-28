package math_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func testDone(f float64) {
	fmt.Sprintf("%f", f)
}

var data [1024]float64

func init() {
	r := rand.New(rand.NewSource(0))
	for i := range data {
		data[i] = r.NormFloat64()
	}
}

func BenchmarkMathAbs(b *testing.B) {
	var f float64
	for i := 0; i < b.N; i++ {
		f += math.Abs(data[i%len(data)])
	}
	b.StopTimer()
	testDone(f)
}
func BenchmarkMathCeil(b *testing.B) {
	var f float64
	for i := 0; i < b.N; i++ {
		f += math.Ceil(data[i%len(data)])
	}
	b.StopTimer()
	testDone(f)
}

func BenchmarkMathExp(b *testing.B) {
	var f float64
	for i := 0; i < b.N; i++ {
		f += math.Exp(data[i%len(data)])
	}
	b.StopTimer()
	testDone(f)
}

func BenchmarkMathExp2(b *testing.B) {
	var f float64
	for i := 0; i < b.N; i++ {
		f += math.Exp2(data[i%len(data)])
	}
	b.StopTimer()
	testDone(f)
}

func BenchmarkMathLog(b *testing.B) {
	var f float64
	for i := 0; i < b.N; i++ {
		f += math.Log(data[i%len(data)])
	}
	b.StopTimer()
	testDone(f)
}
