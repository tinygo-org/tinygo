//go:build baremetal || js || wasm_unknown

package syscall

func Environ() []string {
	env := runtime_envs()
	envCopy := make([]string, len(env))
	copy(envCopy, env)
	return envCopy
}

func Getenv(key string) (value string, found bool) {
	env := runtime_envs()
	for _, keyval := range env {
		// Split at '=' character.
		var k, v string
		for i := 0; i < len(keyval); i++ {
			if keyval[i] == '=' {
				k = keyval[:i]
				v = keyval[i+1:]
			}
		}
		if k == key {
			return v, true
		}
	}
	return "", false
}

func Setenv(key, val string) (err error) {
	// stub for now
	return ENOSYS
}

func Unsetenv(key string) (err error) {
	// stub for now
	return ENOSYS
}

func Clearenv() (err error) {
	// stub for now
	return ENOSYS
}

func runtime_envs() []string
