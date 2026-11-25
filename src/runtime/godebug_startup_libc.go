//go:build (linux && !baremetal && !tinygo.wasm && !nintendoswitch) || darwin || windows

package runtime

func init() {
	if godebugEnvEnabled() {
		godebugStartup = startupEnv("GODEBUG")
	}
}
