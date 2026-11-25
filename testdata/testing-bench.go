package main

import (
	"flag"
	"fmt"
	"testing"
)

func BenchmarkSkip(b *testing.B) {
	b.Skip("skip message")
}

func BenchmarkSkipf(b *testing.B) {
	b.Skipf("skip message %d", 2)
}

func BenchmarkSubSkip(b *testing.B) {
	b.Run("skipped", func(b *testing.B) {
		b.Skip("sub skip message")
	})
}

func BenchmarkFail(b *testing.B) {
	b.Fatal("fail message")
}

type matchString func(pattern, name string) (bool, error)

func (f matchString) MatchString(pattern, name string) (bool, error) {
	return f(pattern, name)
}

func main() {
	testing.Init()
	flag.Set("test.bench", ".")
	m := testing.MainStart(
		matchString(func(string, string) (bool, error) {
			return true, nil
		}),
		nil,
		[]testing.InternalBenchmark{
			{Name: "BenchmarkSkip", F: BenchmarkSkip},
			{Name: "BenchmarkSkipf", F: BenchmarkSkipf},
			{Name: "BenchmarkSubSkip", F: BenchmarkSubSkip},
			{Name: "BenchmarkFail", F: BenchmarkFail},
		},
		nil,
		nil,
	)
	fmt.Println("exit code:", m.Run())
}
