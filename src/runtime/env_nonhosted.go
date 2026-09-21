//go:build baremetal || js || wasm_unknown

package runtime

//go:linkname syscallSetenv syscall.runtimeSetenv
func syscallSetenv(key, value string) {
	entry := key + "=" + value
	for i, e := range env {
		if envKey(e) == key {
			env[i] = entry
			if key == "GODEBUG" {
				godebugSetEnv(value)
			}
			return
		}
	}
	env = append(env, entry)
	if key == "GODEBUG" {
		godebugSetEnv(value)
	}
}

//go:linkname syscallUnsetenv syscall.runtimeUnsetenv
func syscallUnsetenv(key string) {
	for i, e := range env {
		if envKey(e) == key {
			env = append(env[:i], env[i+1:]...)
			if key == "GODEBUG" {
				godebugUnsetEnv()
			}
			return
		}
	}
}

func envKey(entry string) string {
	for i := 0; i < len(entry); i++ {
		if entry[i] == '=' {
			return entry[:i]
		}
	}
	return entry
}
