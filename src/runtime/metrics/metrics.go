// Package metrics is a dummy package that is not yet implemented.

package metrics

type Description struct{}

func All() []Description {
	return nil
}

type Float64Histogram struct{}

type Sample struct{}

func Read(m []Sample) {}

type Value struct{}

func (v Value) Float64() float64 {
	return 0
}
func (v Value) Float64Histogram() *Float64Histogram {
	return nil
}
func (v Value) Kind() ValueKind {
	return ValueKind{}
}
func (v Value) Uint64() uint64 {
	return 0
}

type ValueKind struct{}
