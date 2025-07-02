// Package tinygo contains constants used between the TinyGo compiler and
// runtime.
package tinygo

const (
	PanicStrategyPrint = iota + 1
	PanicStrategyTrap
)

type HashmapAlgorithm uint8

// Constants for hashmap algorithms.
const (
	HashmapAlgorithmBinary HashmapAlgorithm = iota
	HashmapAlgorithmString
	HashmapAlgorithmInterface
)
