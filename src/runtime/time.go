
package runtime

// TODO: use the time package for this.

type Duration uint64

// Use microseconds as the smallest time unit
const (
	Millisecond = Microsecond * 1000
	Second      = Millisecond * 1000
)
