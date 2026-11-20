package runtime

// While gcInhibit is nonzero, allocations grow the heap instead of running the
// garbage collector. Only the collectors that run on demand use this.
var gcInhibit uint32
