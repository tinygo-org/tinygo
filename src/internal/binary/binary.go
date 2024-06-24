// Package binary is a lightweight replacement package for encoding/binary.
package binary

// This file contains small helper functions for working with binary data.

var LittleEndian = littleEndian{}

type littleEndian struct{}

// Encode data like encoding/binary.LittleEndian.Uint16.
func (littleEndian) Uint16(b []byte) uint16 {
	return uint16(b[0]) | uint16(b[1])<<8
}

// Store data like binary.LittleEndian.PutUint16.
func (littleEndian) PutUint16(b []byte, v uint16) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
}

// Append data like binary.LittleEndian.AppendUint16.
func (littleEndian) AppendUint16(b []byte, v uint16) []byte {
	return append(b,
		byte(v),
		byte(v>>8),
	)
}

// Encode data like encoding/binary.LittleEndian.Uint32.
func (littleEndian) Uint32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}
