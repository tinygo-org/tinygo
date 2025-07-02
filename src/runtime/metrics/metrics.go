// Package metrics is a dummy package that is not yet implemented.

package metrics

type Description struct {
	Name        string
	Description string
	Kind        ValueKind
	Cumulative  bool
}

func All() []Description {
	return nil
}

type Float64Histogram struct {
	Counts  []uint64
	Buckets []float64
}

type Sample struct {
	Name  string
	Value Value
}

func Read(m []Sample) {}

type Value struct{}

func (v Value) Float64() float64 {
	return 0
}
func (v Value) Float64Histogram() *Float64Histogram {
	return nil
}
func (v Value) Kind() ValueKind {
	return KindBad
}
func (v Value) Uint64() uint64 {
	return 0
}

type ValueKind int

const (
	KindBad ValueKind = iota
	KindUint64
	KindFloat64
	KindFloat64Histogram
)
