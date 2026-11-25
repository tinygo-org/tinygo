//go:build tinygo.memprofile && !baremetal

package pprof

// enableMemProfile turns on sampling in the allocator. Baremetal test binaries
// cannot write a profile, so it is left out there.
func enableMemProfile()

func init() {
	enableMemProfile()
}
