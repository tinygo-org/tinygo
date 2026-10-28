package builder

import (
	"path/filepath"
)

// The system glibc, linked dynamically.

// glibcStartupJobs returns the startup objects to link, in the order they must
// appear. crt1.o defines _start, which calls __libc_start_main with the main
// that TinyGo's runtime exports; crti.o and crtn.o bracket the .init and .fini
// sections, so crtn.o has to come last.
func glibcStartupJobs(libDir string) []*compileJob {
	var jobs []*compileJob
	for _, name := range []string{"crt1.o", "crti.o", "crtn.o"} {
		jobs = append(jobs, dummyCompileJob(filepath.Join(libDir, name)))
	}
	return jobs
}
