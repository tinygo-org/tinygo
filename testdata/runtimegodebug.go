//go:debug tarinsecurepath=0

package main

import (
	"fmt"
	"strings"
	"syscall"
	_ "unsafe"
)

var defaults, environment string

//go:linkname setUpdate internal/godebug.setUpdate
func setUpdate(update func(defaults, environment string))

func init() {
	setUpdate(func(newDefaults, newEnvironment string) {
		defaults = newDefaults
		environment = newEnvironment
	})
}

func main() {
	printState()
	syscall.Unsetenv("GODEBUG")
	printState()
	syscall.Setenv("GODEBUG", "")
	printState()
	syscall.Setenv("GODEBUG", "tarinsecurepath=1")
	printState()
	syscall.Unsetenv("GODEBUG")
	printState()
	syscall.Setenv("GODEBUG", "tarinsecurepath=1")
	syscall.Clearenv()
	printState()
}

func printState() {
	fmt.Printf("%t %q\n", strings.Contains(defaults, "tarinsecurepath=0"), environment)
}
