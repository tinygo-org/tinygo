// +build output.none

package runtime

// The "none" output drops all logging output (println, panic, ...). This may be
// useful on targets that do not need output logging (for code size, battery
// consumption, or other reasons) or do not have an output at all.

func initOutput() {
}

func putchar(c byte) {
}
