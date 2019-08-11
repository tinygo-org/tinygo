package testing

// B is a type passed to Benchmark functions to manage benchmark timing and to
// specify the number of iterations to run.
//
// TODO: Implement benchmarks. This struct allows test files containing
// benchmarks to compile and run, but will not run the benchmarks themselves.
type B struct {
	N int
}
