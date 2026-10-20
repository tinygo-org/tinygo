//go:build rp2040 || rp2350

package machine_test

import (
	"machine"
	"testing"
)

func TestGetRNGConcurrent(t *testing.T) {
	const workers = 8
	start := make(chan struct{})
	done := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func() {
			<-start
			for j := 0; j < 128; j++ {
				if _, err := machine.GetRNG(); err != nil {
					done <- err
					return
				}
			}
			done <- nil
		}()
	}
	close(start)
	for i := 0; i < workers; i++ {
		if err := <-done; err != nil {
			t.Error(err)
		}
	}
}
