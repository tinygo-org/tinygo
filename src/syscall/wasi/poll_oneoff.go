// +build wasm,wasi

package wasi

//go:wasm-module wasi_unstable
//export poll_oneoff
func Poll_oneoff(in *Subscription_t, out *Event_t, nsubscriptions uint32, nevents *uint32) (errno uint16)

type Eventtype_t = uint8

const (
	Eventtype_t_clock Eventtype_t = 0
	// TODO: __wasi_eventtype_t_fd_read  __wasi_eventtype_t = 1
	// TODO: __wasi_eventtype_t_fd_write __wasi_eventtype_t = 2
)

type (
	// https://github.com/wasmerio/wasmer/blob/1.0.0-alpha3/lib/wasi/src/syscalls/types.rs#L584-L588
	Subscription_t struct {
		UserData uint64
		U        Subscription_u_t
	}

	Subscription_u_t struct {
		Tag Eventtype_t

		// TODO: support fd_read/fd_write event
		U Subscription_clock_t
	}

	// https://github.com/wasmerio/wasmer/blob/1.0.0-alpha3/lib/wasi/src/syscalls/types.rs#L711-L718
	Subscription_clock_t struct {
		UserData  uint64
		ID        uint32
		Timeout   int64
		Precision int64
		Flags     uint16
	}
)

type (
	// https://github.com/wasmerio/wasmer/blob/1.0.0-alpha3/lib/wasi/src/syscalls/types.rs#L191-L198
	Event_t struct {
		UserData  uint64
		Errno     uint16
		EventType Eventtype_t

		// only used for fd_read or fd_write events
		// TODO: support fd_read/fd_write event
		_ struct {
			NBytes uint64
			Flags  uint16
		}
	}
)
