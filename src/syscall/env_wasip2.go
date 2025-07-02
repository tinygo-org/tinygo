//go:build wasip2

package syscall

import (
	"internal/wasi/cli/v0.2.0/environment"
)

var libc_envs map[string]string

func populateEnvironment() {
	libc_envs = make(map[string]string)
	for _, kv := range environment.GetEnvironment().Slice() {
		libc_envs[kv[0]] = kv[1]
	}
}

func Environ() []string {
	var env []string
	for k, v := range libc_envs {
		env = append(env, k+"="+v)
	}
	return env
}

func Getenv(key string) (value string, found bool) {
	value, found = libc_envs[key]
	return
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
	libc_envs[key] = val
	return nil
}

func Unsetenv(key string) (err error) {
	delete(libc_envs, key)
	return nil
}

func Clearenv() {
	clear(libc_envs)
}
