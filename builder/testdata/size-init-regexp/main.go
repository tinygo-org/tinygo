package main

import "regexp"

var identifier = regexp.MustCompile("^[a-z][a-z0-9_]*$")

func main() {
	println(identifier.MatchString("sensor_1"))
	println(identifier.MatchString("1_sensor"))
}
