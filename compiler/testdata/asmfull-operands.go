package main

import "device/arm"

func operand(v uint32) uint32 { return v + 1 }

var operandCount uint32

func firstOperand() uint32 {
	operandCount++
	return operandCount
}

func secondOperand() uint32 {
	operandCount++
	return operandCount
}

func inlineAsmOrdered() uintptr {
	return arm.AsmFull("add {}, {first}, {second}", map[string]interface{}{
		"first":  firstOperand(),
		"second": secondOperand(),
	})
}

func inlineAsm(v uint32) uintptr {
	return arm.AsmFull("mov {}, {value}", map[string]interface{}{
		"value": operand(v),
	})
}

func ordinaryMap(v uint32) interface{} {
	m := map[string]interface{}{"value": v}
	return m["value"]
}

func sharedMap(v uint32) uintptr {
	m := map[string]interface{}{"value": v}
	println(len(m))
	return arm.AsmFull("mov {}, {value}", m)
}
