//go:build !baremetal && (arduino || arduino_nano)

package machine

var I2C0 = &I2C{Bus: 0, PinsSDA: []Pin{PC4}, PinsSCL: []Pin{PC5}}
