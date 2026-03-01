//go:build esp32s3 || esp32c3

package main

import "machine"

const AnalogPin = machine.GPIO4
const led = machine.GPIO4
