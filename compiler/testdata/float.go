package main

// Test converting floats to ints.

func f32tou32(v float32) uint32 {
	return uint32(v)
}

func maxu32f() float32 {
	return float32(^uint32(0))
}

func maxu32tof32() uint32 {
	f := float32(^uint32(0))
	return uint32(f)
}

func inftoi32() (uint32, uint32, int32, int32) {
	inf := 1.0
	inf /= 0.0

	return uint32(inf), uint32(-inf), int32(inf), int32(-inf)
}

func u32tof32tou32(v uint32) uint32 {
	return uint32(float32(v))
}

func f32tou32tof32(v float32) float32 {
	return float32(uint32(v))
}

func f32tou8(v float32) uint8 {
	return uint8(v)
}

func f32toi8(v float32) int8 {
	return int8(v)
}
