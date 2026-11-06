package main

// With a GC that never frees memory, this needs more stack memory than wasm32
// can address unless exited goroutines give their stacks to new ones.

import "sync"

func main() {
	const goroutines = 100000
	const batch = 100
	var wg sync.WaitGroup
	var mu sync.Mutex
	var sum uint64
	for i := 0; i < goroutines; i += batch {
		for j := i; j < i+batch; j++ {
			wg.Add(1)
			go func(j int) {
				defer wg.Done()
				mu.Lock()
				sum += uint64(j)
				mu.Unlock()
			}(j)
		}
		wg.Wait()
	}
	println("sum:", sum)
}
