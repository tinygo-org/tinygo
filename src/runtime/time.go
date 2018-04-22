
package runtime

// TODO: use the time package for this.

// #include <unistd.h>
import "C"

type Duration uint64

// Use microseconds as the smallest time unit
const (
	Microsecond = 1
	Millisecond = Microsecond * 1000
	Second      = Millisecond * 1000
)

func Sleep(d Duration) {
	C.usleep(C.useconds_t(d))
}
