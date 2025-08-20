//go:build !baremetal && (microbit || pca10031 || hw_651)

package machine

var I2C0 = &I2C{Bus: 0}
var I2C1 = &I2C{Bus: 1}
