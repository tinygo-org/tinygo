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
}
