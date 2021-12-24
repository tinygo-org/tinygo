package builder

// This file implements a job runner for the compiler, which runs jobs in
// parallel while taking care of dependencies.

import (
	"container/heap"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"
)

// Set to true to enable logging in the job runner. This may help to debug
// concurrency or performance issues.
const jobRunnerDebug = false

type jobState uint8

const (
	jobStateQueued   jobState = iota // not yet running
	jobStateRunning                  // running
	jobStateFinished                 // finished running
)

// compileJob is a single compiler job, comparable to a single Makefile target.
// It is used to orchestrate various compiler tasks that can be run in parallel
// but that have dependencies and thus have limitations in how they can be run.
type compileJob struct {
	description  string // description, only used for logging
	dependencies []*compileJob
	result       string // result (path)
	run          func(*compileJob) (err error)
	err          error         // error if finished
	duration     time.Duration // how long it took to run this job (only set after finishing)
}

// dummyCompileJob returns a new *compileJob that produces an output without
// doing anything. This can be useful where a *compileJob producing an output is
// expected but nothing needs to be done, for example for a load from a cache.
func dummyCompileJob(result string) *compileJob {
	return &compileJob{
		description: "<dummy>",
		result:      result,
	}
}

// runJobs runs the indicated job and all its dependencies. For every job, all
// the dependencies are run first. It returns the error of the first job that
// fails.
// It runs all jobs in the order of the dependencies slice, depth-first.
// Therefore, if some jobs are preferred to run before others, they should be
// ordered as such in the job dependencies.
func runJobs(job *compileJob, sema chan struct{}) error {
	if sema == nil {
		// Have a default, if the semaphore isn't set. This is useful for
		// tests.
		sema = make(chan struct{}, runtime.NumCPU())
	}
	if cap(sema) == 0 {
		return errors.New("cannot 0 jobs at a time")
	}

	// Create a slice of jobs to run, where all dependencies are run in order.
	jobs := []*compileJob{}
	addedJobs := map[*compileJob]struct{}{}
	var addJobs func(*compileJob)
	addJobs = func(job *compileJob) {
		if _, ok := addedJobs[job]; ok {
			return
		}
		for _, dep := range job.dependencies {
			addJobs(dep)
		}
		jobs = append(jobs, job)
		addedJobs[job] = struct{}{}
	}
	addJobs(job)

	waiting := make(map[*compileJob]map[*compileJob]struct{}, len(jobs))
	dependents := make(map[*compileJob][]*compileJob, len(jobs))
	jidx := make(map[*compileJob]int)
	var ready intHeap
	for i, job := range jobs {
		jidx[job] = i
		if len(job.dependencies) == 0 {
			// This job is ready to run.
			ready.Push(i)
			continue
		}

		// Construct a map for dependencies which the job is currently waiting on.
		waitDeps := make(map[*compileJob]struct{})
		waiting[job] = waitDeps

		// Add the job to the dependents list of each dependency.
		for _, dep := range job.dependencies {
			dependents[dep] = append(dependents[dep], job)
			waitDeps[dep] = struct{}{}
		}
	}

	// Create a channel to accept notifications of completion.
	doneChan := make(chan *compileJob)

	// Send each job in the jobs slice to a worker, taking care of job
	// dependencies.
	numRunningJobs := 0
	var totalTime time.Duration
	start := time.Now()
	for len(ready.IntSlice) > 0 || numRunningJobs != 0 {
		var completed *compileJob
		if len(ready.IntSlice) > 0 {
			select {
			case sema <- struct{}{}:
				// Start a job.
				job := jobs[heap.Pop(&ready).(int)]
				if jobRunnerDebug {
					fmt.Println("## start:   ", job.description)
				}
				go runJob(job, doneChan)
				numRunningJobs++
				continue

			case completed = <-doneChan:
				// A job completed.
			}
		} else {
			// Wait for a job to complete.
			completed = <-doneChan
		}
		numRunningJobs--
		<-sema
		if jobRunnerDebug {
			fmt.Println("## finished:", job.description, "(time "+job.duration.String()+")")
		}
		if completed.err != nil {
			// Wait for any current jobs to finish.
			for numRunningJobs != 0 {
				<-doneChan
				numRunningJobs--
			}

			// The build failed.
			return completed.err
		}

		// Update total run time.
		totalTime += completed.duration

		// Update dependent jobs.
		for _, j := range dependents[completed] {
			wait := waiting[j]
			delete(wait, completed)
			if len(wait) == 0 {
				// This job is now ready to run.
				ready.Push(jidx[j])
				delete(waiting, j)
			}
		}
	}
	if len(waiting) != 0 {
		// There is a dependency cycle preventing some jobs from running.
		return errDependencyCycle{waiting}
	}

	// Some statistics, if debugging.
	if jobRunnerDebug {
		// Total duration of running all jobs.
		duration := time.Since(start)
		fmt.Println("## total:   ", duration)

		// The individual time of each job combined. On a multicore system, this
		// should be lower than the total above.
		fmt.Println("## job sum: ", totalTime)
	}

	return nil
}

type errDependencyCycle struct {
	waiting map[*compileJob]map[*compileJob]struct{}
}

func (err errDependencyCycle) Error() string {
	waits := make([]string, 0, len(err.waiting))
	for j, wait := range err.waiting {
		deps := make([]string, 0, len(wait))
		for dep := range wait {
			deps = append(deps, dep.description)
		}
		sort.Strings(deps)

		waits = append(waits, fmt.Sprintf("\t%s is waiting for [%s]",
			j.description, strings.Join(deps, ", "),
		))
	}
	sort.Strings(waits)
	return "deadlock:\n" + strings.Join(waits, "\n")
}

type intHeap struct {
	sort.IntSlice
}

func (h *intHeap) Push(x interface{}) {
	h.IntSlice = append(h.IntSlice, x.(int))
}

func (h *intHeap) Pop() interface{} {
	x := h.IntSlice[len(h.IntSlice)-1]
	h.IntSlice = h.IntSlice[:len(h.IntSlice)-1]
	return x
}

// runJob runs a compile job and notifies doneChan of completion.
func runJob(job *compileJob, doneChan chan *compileJob) {
	start := time.Now()
	if job.run != nil {
		err := job.run(job)
		if err != nil {
			job.err = err
		}
	}
	job.duration = time.Since(start)
	doneChan <- job
}
