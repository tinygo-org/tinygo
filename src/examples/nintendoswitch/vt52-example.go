// VT52 Demo example
// Same as https://github.com/switchbrew/switch-examples/tree/master/graphics/printing/vt52-demo
package main

import (
	"fmt"
	"github.com/racerxdl/gonx/nx"
)

func main() {
	nx.ConsoleInit(nil)

	// Clear screen and home cursor
	fmt.Printf(nx.ConsoleESC("2J"))

	// Set print co-ordinates
	fmt.Printf(nx.ConsoleESC("10;10H") + "VT52 codes demo")

	// Move cursor up
	fmt.Printf(nx.ConsoleESC("10A") + "Line 0")

	// Move cursor up
	fmt.Printf(nx.ConsoleESC("28D") + "Column 0")

	// Move cursor down
	fmt.Printf(nx.ConsoleESC("19B") + "Line 19")

	// Move cursor right
	fmt.Printf(nx.ConsoleESC("5C") + "Column 20")

	fmt.Println("")

	// Color codes and attributes
	for i := 0; i < 8; i++ {
		fmt.Printf(nx.ConsoleESC("%d;1m")+
			"Default "+
			nx.ConsoleESC("1m")+"Bold "+
			nx.ConsoleESC("7m")+"Reversed "+

			nx.ConsoleESC("0m")+ // Revert Attributes
			nx.ConsoleESC("%d;1m")+

			nx.ConsoleESC("2m")+"Light "+
			nx.ConsoleESC("7m")+"Reversed "+

			nx.ConsoleESC("0m")+ // Revert Attributes
			nx.ConsoleESC("%d;1m")+
			nx.ConsoleESC("4m")+"Underline "+

			nx.ConsoleESC("0m")+ // Revert attributes
			nx.ConsoleESC("%d;1m")+
			nx.ConsoleESC("9m")+"Strikethrough "+

			"\n"+
			nx.ConsoleESC("0m"), i+30, i+30, i+30, i+30)
	}

	for nx.AppletMainLoop() {
		nx.HidScanInput()

		keysDown := nx.HidKeysDown(nx.ControllerP1Auto)

		if keysDown.IsKeyDown(nx.KeyPlus) {
			break
		}

		nx.ConsoleUpdate(nil)
	}

	nx.ConsoleExit(nil)
}
