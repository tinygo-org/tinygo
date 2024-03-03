// //go:build tinygo.wasm

package math_test

import (
	"math"
	"math/rand"
	"testing"
)

var tested float64

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
		f = math.Abs(data[i%len(data)])
	}
	tested = f
}
func BenchmarkMathCeil(b *testing.B) {
	var f float64
	for i := 0; i < b.N; i++ {
		f = math.Ceil(data[i%len(data)])
	}
	tested = f
}
func BenchmarkMathExp(b *testing.B) {
	var f float64
	for i := 0; i < b.N; i++ {
		f = math.Exp(data[i%len(data)])
	}
	tested = f
}

func BenchmarkMathExp2(b *testing.B) {
	var f float64
	for i := 0; i < b.N; i++ {
		f = math.Exp2(data[i%len(data)])
	}
	tested = f
}

func BenchmarkMathLog(b *testing.B) {
	var f float64
	for i := 0; i < b.N; i++ {
		f = math.Log(data[i%len(data)])
	}
	tested = f
}
