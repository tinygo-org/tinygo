//go:build !baremetal && hifive1b

package machine

var I2C0 = &I2C{Bus: 0, PinsSDA: []Pin{P12}, PinsSCL: []Pin{P13}}
