package runtime_test

import (
	"math"
	"testing"
	_ "unsafe"
)

func TestFloatMinMax32(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		x   float32
		y   float32
		min float32
		max float32
	}{
		{
			x:   0,
			y:   0,
			min: 0,
			max: 0,
		},
		{
			x:   -12,
			y:   2,
			min: -12,
			max: 2,
		},
		{
			x:   2,
			y:   -12,
			min: -12,
			max: 2,
		},
		{
			x:   float32(math.Copysign(0, -1)),
			y:   0,
			min: float32(math.Copysign(0, -1)),
			max: 0,
		},
		{
			x:   0,
			y:   float32(math.Copysign(0, -1)),
			min: float32(math.Copysign(0, -1)),
			max: 0,
		},
		{
			x:   float32(math.Inf(-1)),
			y:   float32(math.Inf(1)),
			min: float32(math.Inf(-1)),
			max: float32(math.Inf(1)),
		},
		{
			x:   math.MaxFloat32,
			y:   math.SmallestNonzeroFloat32,
			min: math.SmallestNonzeroFloat32,
			max: math.MaxFloat32,
		},
		{
			x:   math.Float32frombits(float32PositiveNaN),
			y:   0,
			min: math.Float32frombits(float32PositiveNaN),
			max: math.Float32frombits(float32PositiveNaN),
		},
		{
			x:   0,
			y:   math.Float32frombits(float32PositiveNaN),
			min: math.Float32frombits(float32PositiveNaN),
			max: math.Float32frombits(float32PositiveNaN),
		},
		{
			x:   math.Float32frombits(float32PositiveNaN),
			y:   math.Float32frombits(float32PositiveNaN),
			min: math.Float32frombits(float32PositiveNaN),
			max: math.Float32frombits(float32PositiveNaN),
		},
		{
			x:   math.Float32frombits(float32NegativeNaN),
			y:   0,
			min: math.Float32frombits(float32NegativeNaN),
			max: math.Float32frombits(float32NegativeNaN),
		},
		{
			x:   0,
			y:   math.Float32frombits(float32NegativeNaN),
			min: math.Float32frombits(float32NegativeNaN),
			max: math.Float32frombits(float32NegativeNaN),
		},
		{
			x:   math.Float32frombits(float32NegativeNaN),
			y:   math.Float32frombits(float32NegativeNaN),
			min: math.Float32frombits(float32NegativeNaN),
			max: math.Float32frombits(float32NegativeNaN),
		},
	} {
		if min := minimumFloat32(c.x, c.y); math.Float32bits(min) != math.Float32bits(c.min) {
			t.Errorf("minimumFloat32(%f, %f) = %f (expected %f)", c.x, c.y, min, c.min)
		}
		if max := maximumFloat32(c.x, c.y); math.Float32bits(max) != math.Float32bits(c.max) {
			t.Errorf("maximumFloat32(%f, %f) = %f (expected %f)", c.x, c.y, max, c.max)
		}
	}
}

const (
	// float32PositiveNaN is the smallest positive NaN value for a float32.
	float32PositiveNaN = 0x7FC00001
	// float32NegativeNaN is the smallest negative NaN value for a float32.
	float32NegativeNaN = 0xFFC00001
)

//go:linkname minimumFloat32 runtime.minimumFloat32
func minimumFloat32(x, y float32) float32

//go:linkname maximumFloat32 runtime.maximumFloat32
func maximumFloat32(x, y float32) float32

func TestFloatMinMax64(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		x   float64
		y   float64
		min float64
		max float64
	}{
		{
			x:   0,
			y:   0,
			min: 0,
			max: 0,
		},
		{
			x:   -12,
			y:   2,
			min: -12,
			max: 2,
		},
		{
			x:   2,
			y:   -12,
			min: -12,
			max: 2,
		},
		{
			x:   math.Copysign(0, -1),
			y:   0,
			min: math.Copysign(0, -1),
			max: 0,
		},
		{
			x:   0,
			y:   math.Copysign(0, -1),
			min: math.Copysign(0, -1),
			max: 0,
		},
		{
			x:   math.Inf(-1),
			y:   math.Inf(1),
			min: math.Inf(-1),
			max: math.Inf(1),
		},
		{
			x:   math.MaxFloat64,
			y:   math.SmallestNonzeroFloat64,
			min: math.SmallestNonzeroFloat64,
			max: math.MaxFloat64,
		},
		{
			x:   math.Float64frombits(float64PositiveNaN),
			y:   0,
			min: math.Float64frombits(float64PositiveNaN),
			max: math.Float64frombits(float64PositiveNaN),
		},
		{
			x:   0,
			y:   math.Float64frombits(float64PositiveNaN),
			min: math.Float64frombits(float64PositiveNaN),
			max: math.Float64frombits(float64PositiveNaN),
		},
		{
			x:   math.Float64frombits(float64PositiveNaN),
			y:   math.Float64frombits(float64PositiveNaN),
			min: math.Float64frombits(float64PositiveNaN),
			max: math.Float64frombits(float64PositiveNaN),
		},
		{
			x:   math.Float64frombits(float64NegativeNaN),
			y:   0,
			min: math.Float64frombits(float64NegativeNaN),
			max: math.Float64frombits(float64NegativeNaN),
		},
		{
			x:   0,
			y:   math.Float64frombits(float64NegativeNaN),
			min: math.Float64frombits(float64NegativeNaN),
			max: math.Float64frombits(float64NegativeNaN),
		},
		{
			x:   math.Float64frombits(float64NegativeNaN),
			y:   0,
			min: math.Float64frombits(float64NegativeNaN),
			max: math.Float64frombits(float64NegativeNaN),
		},
	} {
		if min := minimumFloat64(c.x, c.y); math.Float64bits(min) != math.Float64bits(c.min) {
			t.Errorf("minimumFloat64(%f, %f) = %f (expected %f)", c.x, c.y, min, c.min)
		}
		if max := maximumFloat64(c.x, c.y); math.Float64bits(max) != math.Float64bits(c.max) {
			t.Errorf("maximumFloat64(%f, %f) = %f (expected %f)", c.x, c.y, max, c.max)
		}
	}
}

const (
	// float64PositiveNaN is the smallest positive NaN value for a float64.
	float64PositiveNaN = 0x7FF8000000000001
	// float64NegativeNaN is the smallest negative NaN value for a float64.
	float64NegativeNaN = 0xFFF8000000000001
)

//go:linkname minimumFloat64 runtime.minimumFloat64
func minimumFloat64(x, y float64) float64

//go:linkname maximumFloat64 runtime.maximumFloat64
func maximumFloat64(x, y float64) float64
