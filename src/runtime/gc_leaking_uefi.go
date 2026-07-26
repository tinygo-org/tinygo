//go:build gc.leaking && uefi

package runtime

import _ "unsafe"

//go:export tinygo_scanstack
func tinygo_scanstack() {
}
