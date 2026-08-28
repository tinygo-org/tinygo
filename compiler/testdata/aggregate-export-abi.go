package main

type exportedAggregateParamMethod struct{}

//export exportedAggregateParamMethodCall
func (exportedAggregateParamMethod) call(value [600]int32, other [600]int32) int32 {
	return value[0] + other[len(other)-1]
}

//export exportedOversizedAggregate
func exportedOversizedAggregate(value [1001]int32) {
}

func exerciseExportedAggregateMethods() {
	var paramMethod interface {
		call([600]int32, [600]int32) int32
	} = exportedAggregateParamMethod{}
	paramMethod.call([600]int32{}, [600]int32{})
}
