package main

var crc8Table = makeCRC8Table()

// makeCRC8Table runs at startup until the interp loop limit is raised.
// See https://github.com/tinygo-org/tinygo/issues/6045.
func makeCRC8Table() [256]uint8 {
	var table [256]uint8
	for i := range table {
		crc := uint8(i)
		for range 8 {
			if crc&0x80 != 0 {
				crc = crc<<1 ^ 0x07
			} else {
				crc <<= 1
			}
		}
		table[i] = crc
	}
	return table
}

func crc8(data []byte) uint8 {
	var crc uint8
	for _, b := range data {
		crc = crc8Table[crc^b]
	}
	return crc
}

func main() {
	println(crc8([]byte("hello")))
}
