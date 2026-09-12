package main

type exportedAggregateParamMethod struct{}

//export exportedAggregateParamMethodCall
func (exportedAggregateParamMethod) call(value [600]int32, other [600]int32) int32 {
	return value[0] + other[len(other)-1]
}

//export exportedOversizedAggregate
func exportedOversizedAggregate(value [1001]int32) {
}

type exportedAggregateResultMethod struct{}

//export exportedAggregateResultMethodCall
func (exportedAggregateResultMethod) call() [1025]int32 {
	return [1025]int32{}
}

type exportedLargeReceiver [600]int32

//export exportedLargeReceiverCall
func (receiver exportedLargeReceiver) call(value [399]int32) int32 {
	return receiver[0] + value[0]
}

func exerciseExportedAggregateMethods() {
	var paramMethod interface {
		call([600]int32, [600]int32) int32
	} = exportedAggregateParamMethod{}
	paramMethod.call([600]int32{}, [600]int32{})

	var resultMethod interface {
		call() [1025]int32
	} = exportedAggregateResultMethod{}
	resultMethod.call()

	var receiverMethod interface {
		call([399]int32) int32
	} = exportedLargeReceiver{}
	receiverMethod.call([399]int32{})
}
