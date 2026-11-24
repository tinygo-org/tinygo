package runtime

// MemProfileRate is the average number of bytes allocated per sample in the
// memory profile. Only -gc=precise and -gc=conservative support it.
var MemProfileRate int = 512 * 1024

// A MemProfileRecord describes the live objects allocated by a particular
// call sequence (stack trace).
type MemProfileRecord struct {
	AllocBytes, FreeBytes     int64       // number of bytes allocated, freed
	AllocObjects, FreeObjects int64       // number of objects allocated, freed
	Stack0                    [32]uintptr // stack trace for this record; ends at first 0 entry
}

// InUseBytes returns the number of bytes in use (AllocBytes - FreeBytes).
func (r *MemProfileRecord) InUseBytes() int64 { return r.AllocBytes - r.FreeBytes }

// InUseObjects returns the number of objects in use (AllocObjects - FreeObjects).
func (r *MemProfileRecord) InUseObjects() int64 {
	return r.AllocObjects - r.FreeObjects
}

// Stack returns the stack trace associated with the record, a prefix of
// r.Stack0.
func (r *MemProfileRecord) Stack() []uintptr {
	for i, v := range r.Stack0 {
		if v == 0 {
			return r.Stack0[0:i]
		}
	}
	return r.Stack0[0:]
}

// MemProfile returns the memory profile as of the last GC, as upstream does.
// See https://pkg.go.dev/runtime#MemProfile.
func MemProfile(p []MemProfileRecord, inuseZero bool) (n int, ok bool) {
	return memProfile(p, inuseZero)
}

// memProfileEnabled is set by the init of runtime/pprof, which interp runs at
// build time. Without runtime/pprof the profiling code is optimized away.
var memProfileEnabled bool

//go:linkname pprof_enableMemProfile runtime/pprof.enableMemProfile
func pprof_enableMemProfile() {
	memProfileEnabled = true
}
