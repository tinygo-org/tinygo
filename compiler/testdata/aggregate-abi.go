package main

type aggregateValue [600]int32

// The largest aggregate that is passed directly.
type budgetAggregate [16]int32

type splitAggregate [17]int32

type moderateAggregate [400]int32

type interfaceBoundaryAggregate [998]int32

//go:noinline
func readDirectAggregate(x budgetAggregate, y splitAggregate) int32 {
	return x[0] + y[0]
}

// The context and 62 aggregates are 993 parameters, 7 short of the limit.
//
//go:noinline
func readLimitAggregates(a0, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15, a16, a17, a18, a19, a20, a21, a22, a23, a24, a25, a26, a27, a28, a29, a30, a31, a32, a33, a34, a35, a36, a37, a38, a39, a40, a41, a42, a43, a44, a45, a46, a47, a48, a49, a50, a51, a52, a53, a54, a55, a56, a57, a58, a59, a60, a61 budgetAggregate, b0, b1, b2, b3, b4, b5, b6 int32) int32 {
	return a0[0] + a61[0] + b6
}

//go:noinline
func readBoundaryAggregates(a0, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15, a16, a17, a18, a19, a20, a21, a22, a23, a24, a25, a26, a27, a28, a29, a30, a31, a32, a33, a34, a35, a36, a37, a38, a39, a40, a41, a42, a43, a44, a45, a46, a47, a48, a49, a50, a51, a52, a53, a54, a55, a56, a57, a58, a59, a60, a61, a62 budgetAggregate) int32 {
	return a0[0] + a62[0]
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
func readResultBudget(a0, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15, a16, a17, a18, a19, a20, a21, a22, a23, a24, a25, a26, a27, a28, a29, a30, a31, a32, a33, a34, a35, a36, a37, a38, a39, a40, a41, a42, a43, a44, a45, a46, a47, a48, a49, a50, a51, a52, a53, a54, a55, a56, a57, a58, a59, a60, a61 budgetAggregate, b0, b1, b2, b3, b4, b5, b6 int32) (int32, int32) {
	return a0[0], a61[0] + b6
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
