package main

func makeLargeElementSlice(len int) [][32 << 20]byte {
	return make([][32 << 20]byte, len)
}
