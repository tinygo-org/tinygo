package main

type register struct {
	reg, val uint8
}

type config struct {
	address       uint16
	width, height int16
	init          []register
}

var display = config{
	address: 0x3C,
	width:   128,
	height:  64,
	init: []register{
		{0xAE, 0x00},
		{0xD5, 0x80},
		{0xA8, 0x3F},
		{0xD3, 0x00},
		{0x8D, 0x14},
		{0x20, 0x00},
		{0xA1, 0x00},
		{0xC8, 0x00},
		{0x81, 0xCF},
		{0xAF, 0x00},
	},
}

func main() {
	println(display.address, display.width, display.height)
	for _, r := range display.init {
		println(r.reg, r.val)
	}
}
