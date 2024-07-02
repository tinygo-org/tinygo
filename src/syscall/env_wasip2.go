//go:build wasip2

package syscall

func Environ() []string {
	var env []string
	for k, v := range libc_envs {
		env = append(env, k+"="+v)
	}
	return env
}
