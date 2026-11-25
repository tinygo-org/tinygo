package builder

import (
	"errors"
	"testing"
)

func TestRunJobsRejectsZeroCapacitySemaphore(t *testing.T) {
	for _, tc := range []struct {
		name string
		sema chan struct{}
	}{
		{name: "nil"},
		{name: "unbuffered", sema: make(chan struct{})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := runJobs(dummyCompileJob(""), tc.sema); err == nil || err.Error() != "cannot run 0 jobs at a time" {
				t.Fatalf("runJobs returned %v, want cannot run 0 jobs at a time", err)
			}
		})
	}
}

func TestRunJobsReleasesSemaphoreAfterFailure(t *testing.T) {
	sema := make(chan struct{}, 2)
	otherStarted := make(chan struct{})
	finishOther := make(chan struct{})
	want := errors.New("expected failure")

	failing := &compileJob{
		run: func(*compileJob) error {
			<-otherStarted
			return want
		},
	}
	other := &compileJob{
		run: func(*compileJob) error {
			close(otherStarted)
			<-finishOther
			return nil
		},
	}

	// Let the other job finish only after runJobs has handled the failure.
	// Both permits are taken until then, so this send blocks.
	go func() {
		<-otherStarted
		sema <- struct{}{}
		<-sema
		close(finishOther)
	}()

	job := &compileJob{dependencies: []*compileJob{failing, other}}
	if err := runJobs(job, sema); err != want {
		t.Fatalf("runJobs returned %v, want %v", err, want)
	}
	if got := len(sema); got != 0 {
		t.Fatalf("runJobs retained %d semaphore permits after all jobs finished", got)
	}
}
