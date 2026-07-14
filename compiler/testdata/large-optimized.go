package main

type largeOptimizedValue [1025]byte

type mixedLargeOptimizedValue struct {
	value [1025]byte
	any   any
}

func makeLargeOptimizedValue() largeOptimizedValue {
	return largeOptimizedValue{}
}

func readLargeOptimizedValue(value largeOptimizedValue) byte {
	return value[len(value)-1]
}

func readMixedLargeOptimizedValue(value mixedLargeOptimizedValue) byte {
	return value.value[len(value.value)-1]
}
