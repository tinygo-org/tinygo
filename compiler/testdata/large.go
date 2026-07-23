package main

type largeValue [1025]byte

type largeStruct struct {
	value [1025]byte
	ptr   *byte
}

type largeInterface interface {
	makeLargeValue() largeValue
	readLargeValue(largeValue) byte
}

type largeReceiver largeValue

func makeLargeValue(value byte) largeValue {
	var result largeValue
	result[len(result)-1] = value
	return result
}

func makeZeroLargeValue() largeValue {
	return largeValue{}
}

func passZeroLargeValue() byte {
	return readLargeValue(largeValue{})
}

func readLargeValue(value largeValue) byte {
	return value[len(value)-1]
}

func useLargeValue() byte {
	return readLargeValue(makeLargeValue(42))
}

func useLargeFunctionValue(fn func(largeValue) byte) byte {
	return fn(makeLargeValue(42))
}

func (receiver largeReceiver) makeLargeValue() largeValue {
	return largeValue(receiver)
}

func (receiver largeReceiver) readLargeValue(value largeValue) byte {
	return value[len(value)-1]
}

func useLargeInterface(value largeInterface) byte {
	return value.readLargeValue(value.makeLargeValue())
}

func deferLargeValue(value largeValue) {
	defer readLargeValue(value)
}

func goLargeValue(value largeValue) {
	go readLargeValue(value)
}

func makeLargeResults(value byte) (largeValue, byte) {
	return makeLargeValue(value), value
}

func makeTwoLargeResults(value byte) (largeValue, largeValue) {
	return makeLargeValue(value), makeLargeValue(value + 1)
}

func makeMixedLargeResults(value byte) (largeValue, byte, largeValue) {
	return makeLargeValue(value), value + 1, makeLargeValue(value + 2)
}

func chooseLargeValue(flag bool) largeValue {
	value := makeLargeValue(1)
	if flag {
		value = makeLargeValue(42)
	}
	return value
}

func makePointerLargeValue(value *byte) largeStruct {
	return largeStruct{ptr: value}
}

func assertLargeValue(value any) byte {
	large, ok := value.(largeValue)
	if !ok {
		return 0
	}
	return large[len(large)-1]
}

func useLargeMap(key, value largeValue) byte {
	values := map[largeValue]largeValue{key: value}
	result, ok := values[key]
	if !ok {
		return 0
	}
	return result[len(result)-1]
}

func useLargeChannel(ch chan largeValue, value largeValue) byte {
	ch <- value
	result, ok := <-ch
	if !ok {
		return 0
	}
	return result[len(result)-1]
}

func selectLargeChannel(ch chan largeValue, value largeValue) byte {
	select {
	case ch <- value:
		return 0
	case result := <-ch:
		return result[len(result)-1]
	}
}
