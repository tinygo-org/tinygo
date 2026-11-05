package main

type byteSlice []byte
type textString string

type stringField struct {
	text string
	n    int
}

type sliceField struct {
	data []byte
	n    int
}

type integerFields struct {
	len int
	cap int
}

func sliceArgs(s []byte, n int) int {
	return len(s) + cap(s) + n
}

func stringArg(s string) int {
	return len(s)
}

func namedArgs(s byteSlice, text textString) int {
	return len(s) + cap(s) + len(text)
}

func nestedArg(s stringField) int {
	return len(s.text) + s.n
}

func unexpandedArg(s sliceField) int {
	return len(s.data) + s.n
}

func ordinaryFields(s integerFields) int {
	return s.len + s.cap
}

func indirectResult(s []byte) [1025]byte {
	return [1025]byte{}
}

func (s byteSlice) size() int {
	return len(s)
}

func receiverArg(s byteSlice) int {
	return s.size()
}

//export exportedArgs
func exportedArgs(s []byte, text string) int {
	return len(s) + cap(s) + len(text)
}

func negativeLength(s []byte) bool {
	return len(s) < 0
}

func negativeCapacity(s []byte) bool {
	return cap(s) < 0
}

func negativeString(s string) bool {
	return len(s) < 0
}

func negativeInt(n int) bool {
	return n < 0
}

func countedSum(s []byte) int {
	sum := 0
	for i := 0; i < len(s); i++ {
		sum += int(s[i])
	}
	return sum
}

func rangeSum(s []byte) int {
	sum := 0
	for _, v := range s {
		sum += int(v)
	}
	return sum
}

func index(s []byte, i int) byte {
	return s[i]
}

func tail(s []byte, i int) []byte {
	return s[i:]
}
