//go:build runtime_hexdump

package runtime

import "unsafe"

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func Hexdump(p []byte) {
	var hd hexdumper
	var completedBytes int
	for i := 0; i < len(p); i += 16 {
		m := min(len(p), i+16)
		hd.format(p[i:m])
		completedBytes += m - i
		hd.offset += uint(m - i)
	}
}

var hex = []byte("0123456789abcdef")

type hexdumper struct {
	offset uint
	line   [83]byte
}

func (hd *hexdumper) format(buf []byte) {
	// addr:_(hex dump)+spaces+space+bar+chars+bar+newline
	line := hd.line[:]
	ptr := 0

	offs := hd.offset
	ptr += 8

	line[ptr] = ':'
	ptr--

	for offs > 0 {
		line[ptr] = hex[offs&0x0f]
		ptr--
		offs >>= 4
	}

	for ptr >= 0 {
		line[ptr] = '0'
		ptr--
	}

	ptr = 9
	line[ptr] = ' '
	ptr++

	for i, b := range buf {
		if i%4 == 0 {
			line[ptr] = ' '
			ptr++
		}

		line[ptr] = hex[b>>4]
		ptr++
		line[ptr] = hex[b&0x0f]
		ptr++
		line[ptr] = ' '
		ptr++

	}

	// fill in rest of line
	for i := len(buf); i < 16; i++ {
		if i%4 == 0 {
			line[ptr] = ' '
			ptr++
		}

		line[ptr] = ' '
		ptr++
		line[ptr] = ' '
		ptr++
		line[ptr] = ' '
		ptr++
	}

	line[ptr] = ' '
	ptr++
	line[ptr] = ' '
	ptr++
	line[ptr] = '|'
	ptr++

	for _, v := range buf {
		if v > 32 && v < 127 {
			line[ptr] = v
		} else {
			line[ptr] = '.'
		}
		ptr++
	}

	line[ptr] = '|'
	ptr++

	line[ptr] = '\n'
	ptr++

	print(unsafe.String(&line[0], ptr))
}
