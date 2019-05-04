package main

func main() {
	// sanity
	println(3.14159265358979323846)

	// float64
	f64 := float64(2) / float64(3)
	println(f64)
	println(f64 + 1.0)
	println(f64 - 1.0)
	println(f64 * 2.0)
	println(f64 / 2.0)

	// float32
	f32 := float32(2) / float32(3)
	println(f32)
	println(f32 + 1.0)
	println(f32 - 1.0)
	println(f32 * 2.0)
	println(f32 / 2.0)

	// casting
	println(float32(f64))
	println(float64(f32))

	// float -> int
	var f1 float32 = 3.3
	var f2 float32 = 5.7
	var f3 float32 = -2.3
	var f4 float32 = -11.8
	println(int32(f1), int32(f2), int32(f3), int32(f4))

	// int -> float
	var i1 int32 = 53
	var i2 int32 = -8
	var i3 uint32 = 20
	println(float32(i1), float32(i2), float32(i3))

	// complex64
	c64 := complex(f32, 1.2)
	println(c64)
	println(real(c64))
	println(imag(c64))

	// complex128
	c128 := complex(f64, -2.0)
	println(c128)
	println(real(c128))
	println(imag(c128))

	// untyped complex
	println(2 + 1i)
	println(complex(2, -2))

	// cast complex
	println(complex64(c128))
	println(complex128(c64))

	// binops on complex numbers
	c64 = 5+2i
	println("complex64 add: ", c64 + -3+8i)
	println("complex64 sub: ", c64 - -3+8i)
	c128 = -5+2i
	println("complex128 add:", c128 + 2+6i)
	println("complex128 sub:", c128 - 2+6i)
}
