// This is a echo console running on the os.Stdin and os.Stdout.
// Stdin and os.Stdout are connected to machine.Serial in the baremetal target.
//
// Serial can be switched with the -serial option as follows
// 1. tinygo flash -target wioterminal -serial usb examples/echo2
// 2. tinygo flash -target wioterminal -serial uart examples/echo2
//
// This example will also work with standard Go.
package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Printf("Echo console enabled. Type something then press enter:\r\n")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		msg := ""
		fmt.Scanf("%s\n", &msg)
		fmt.Printf("You typed (scanf) : %s\r\n", msg)

		if scanner.Scan() {
			fmt.Printf("You typed (scanner) : %s\r\n", scanner.Text())
		}
	}
}
