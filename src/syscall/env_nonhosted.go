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
	if len(key) == 0 {
		return EINVAL
	}
	for i := 0; i < len(key); i++ {
		if key[i] == '=' || key[i] == 0 {
			return EINVAL
		}
	}
	for i := 0; i < len(val); i++ {
		if val[i] == 0 {
			return EINVAL
		}
	}
	runtimeSetenv(key, val)
	return nil
}

func Unsetenv(key string) (err error) {
	runtimeUnsetenv(key)
	return nil
}

func Clearenv() (err error) {
	for _, s := range Environ() {
		for j := 0; j < len(s); j++ {
			if s[j] == '=' {
				Unsetenv(s[:j])
				break
			}
		}
	}
	return nil
}

func runtime_envs() []string
