package main

var values = makeValues()

func makeValues() []byte {
	var values []byte
	for i := 0; i < 32; i++ {
		values = append(values, byte(i))
	}
	return values
}

func main() {
	for _, value := range values {
		println(value)
	}
}
