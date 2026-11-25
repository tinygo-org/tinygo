//go:build esp8266

package interrupt

// In returns whether the system is currently in an interrupt.
// esp8266 does not currently support interrupt dispatch to Go functions,
// so this always returns false.
func In() bool {
	return false
}
