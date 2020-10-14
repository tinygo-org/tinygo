// +build output.custom

package runtime

// The "custom" output allows setting a custom function for putchar, for full
// control over where output goes.

var outputFunc func(c byte)

func initOutput() {
}

func putchar(c byte) {
	if outputFunc != nil {
		outputFunc(c)
	}
}

// SetOutput sets a custom function that may be used as an output. This may be
// anything, for example it can be used to redirect output to ARM semihosting if
// needed while debugging.
func SetOutput(output func(byte)) {
	outputFunc = output
}
