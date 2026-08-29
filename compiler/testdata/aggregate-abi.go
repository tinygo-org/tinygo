package main

type aggregateValue [600]int32

type directAggregate [499]int32

type boundaryAggregate [500]int32

type moderateAggregate [400]int32

type interfaceBoundaryAggregate [998]int32

//go:noinline
func readDirectAggregates(x, y directAggregate) int32 {
	return x[0] + y[len(y)-1]
}

//go:noinline
func readLimitAggregates(x, y directAggregate) (int32, int32) {
	return x[0], y[0]
}

//go:noinline
func readBoundaryAggregates(x, y boundaryAggregate) int32 {
	return x[0] + y[len(y)-1]
}

//go:noinline
func readAggregates(x, y aggregateValue) int32 {
	return x[0] + y[len(y)-1]
}

//go:noinline
func readSingleAggregate(value aggregateValue) int32 {
	return value[0]
}

//go:noinline
func readThreeAggregates(x, y, z moderateAggregate) int32 {
	return x[0] + y[0] + z[0]
}

//go:noinline
func readResultBudget(x directAggregate, y boundaryAggregate) (int32, int32) {
	return x[0], y[0]
}

func selectAggregate(cond bool, x, y aggregateValue) int32 {
	selected := x
	if cond {
		selected = y
	}
	return readSingleAggregate(selected)
}

func callAggregates(x, y aggregateValue) int32 {
	return readAggregates(x, y)
}

func callAggregateFunction(fn func(aggregateValue, aggregateValue) int32, x, y aggregateValue) int32 {
	return fn(x, y)
}

type aggregateReceiver struct{}

type aggregateInterface interface {
	read(aggregateValue, aggregateValue) int32
}

type boundaryInterface interface {
	readBoundary(interfaceBoundaryAggregate) int32
}

//go:noinline
func (aggregateReceiver) read(x, y aggregateValue) int32 {
	return x[0] + y[len(y)-1]
}

func callAggregateMethod(receiver aggregateReceiver, x, y aggregateValue) int32 {
	return receiver.read(x, y)
}

func callBoundAggregateMethod(receiver aggregateReceiver, x, y aggregateValue) int32 {
	method := receiver.read
	return method(x, y)
}

func callAggregateInterface(receiver aggregateInterface, x, y aggregateValue) int32 {
	return receiver.read(x, y)
}

func (aggregateReceiver) readBoundary(value interfaceBoundaryAggregate) int32 {
	return value[0]
}

func callBoundaryInterface(receiver boundaryInterface, value interfaceBoundaryAggregate) int32 {
	return receiver.readBoundary(value)
}

func deferAggregates(x, y aggregateValue) {
	defer readAggregates(x, y)
}

func deferAggregateFunction(fn func(aggregateValue, aggregateValue) int32, x, y aggregateValue) {
	defer fn(x, y)
}

func goAggregates(x, y aggregateValue) {
	go readAggregates(x, y)
}

func goAggregateFunction(fn func(aggregateValue, aggregateValue) int32, x, y aggregateValue) {
	go fn(x, y)
}

//export readAggregateExport
func readAggregateExport(value aggregateValue) int32 {
	return value[0]
}
