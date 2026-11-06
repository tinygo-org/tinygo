//go:build (linux && !baremetal && !wasm_unknown && !wasip2 && !nintendoswitch) || darwin || windows

package runtime

func init() {
	if godebugEnvEnabled() {
		godebugStartup = startupEnv("GODEBUG")
	}
}
