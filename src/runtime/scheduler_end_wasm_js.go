//go:build wasm && !wasi

package runtime

func postMain() {
	proc_exit(0)
}
