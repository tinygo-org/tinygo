
package main

const SIX = 6

func main() {
	println("Hello world from Go!")
	println("The answer is:", calculateAnswer())
}

func calculateAnswer() int {
	seven := 7
	return SIX * seven
}
