package cases

import (
	"math"
	"strconv"
	"strings"
)

// The explicit conversions stop the host compiler from fusing a multiply and
// an add, which the soft float code on the device does not do.

func init() {
	register("float64-basic", 0, func() string {
		r := rng(20)
		var s float64
		acc := uint64(0)
		for i := 0; i < 500; i++ {
			x := float64(int32(r.next())) / 65536.0
			y := float64(int32(r.next()))/1024.0 + 0.5
			s += float64(x * y)
			s -= float64(x / (y + 1e6))
			s = float64(s*0.999) + math.Sqrt(math.Abs(x))
			acc = acc*31 + math.Float64bits(s)
		}
		return hex64(acc) + " " + hex64(math.Float64bits(s))
	})

	register("float32-basic", 0, func() string {
		r := rng(21)
		var s float32
		acc := uint64(0)
		for i := 0; i < 500; i++ {
			x := float32(int32(r.next())) / 65536.0
			y := float32(int32(r.next()))/1024.0 + 0.5
			s += float32(x * y)
			s -= float32(x / (y + 1e6))
			s = float32(s*0.999) + float32(math.Sqrt(math.Abs(float64(x))))
			acc = acc*31 + uint64(math.Float32bits(s))
		}
		return hex64(acc) + " " + hex64(uint64(math.Float32bits(s)))
	})

	register("math-funcs", 8, func() string {
		var out []string
		xs := []float64{0.1, 0.5, 1, 1.5, 2.5, 3.14159, 7, 10.25, 100.5, 0.001}
		for _, x := range xs {
			for _, v := range []float64{
				math.Sin(x), math.Cos(x), math.Tan(x), math.Exp(x / 4), math.Log(x),
				math.Pow(x, 1.7), math.Atan2(x, 3), math.Sinh(x / 5), math.Cbrt(x),
			} {
				out = append(out, hex64(math.Float64bits(v)))
			}
		}
		return strings.Join(out, " ")
	})

	register("strconv", 0, func() string {
		r := rng(22)
		var sb strings.Builder
		for i := 0; i < 60; i++ {
			f := math.Float64frombits(uint64(r.next())<<32|uint64(r.next())) / 1e30
			if math.IsNaN(f) || math.IsInf(f, 0) {
				f = float64(i)
			}
			sb.WriteString(strconv.FormatFloat(f, 'g', -1, 64))
			sb.WriteString(strconv.FormatFloat(f, 'e', 6, 64))
			sb.WriteString(strconv.FormatFloat(f, 'f', 3, 32))
			back, err := strconv.ParseFloat(strconv.FormatFloat(f, 'g', -1, 64), 64)
			if err != nil || back != f {
				sb.WriteString("roundtrip-failed")
			}
		}
		for base := 2; base <= 36; base++ {
			v := int64(int32(r.next()))
			s := strconv.FormatInt(v, base)
			back, err := strconv.ParseInt(s, base, 64)
			if err != nil || back != v {
				sb.WriteString("int-roundtrip-failed")
			}
			sb.WriteString(s)
		}
		s := sb.String()
		return strconv.Itoa(len(s)) + " " + hex64(fnv64(s))
	})
	needsHeap(8 << 10)
}
