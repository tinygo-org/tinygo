package main

// Generic type alias (requires go 1.24+, or GOEXPERIMENT=aliastypeparams in go 1.23).
type Set[T comparable] = map[T]struct{}

func main() {
	s := make(Set[string])
	s["hello"] = struct{}{}
	println(len(s))
}
