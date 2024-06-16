//go:build tinygo

package syscall

func Unsetenv(key string) error {
	println("syscall.Unsetenv not implemented", key)
	return EOPNOTSUPP
}

func Getenv(key string) (value string, found bool) {
	println("syscall.Getenv not implemented", key)
	return value, false
}

func Setenv(key, value string) error {
	println("syscall.Setenv not implemented", key, value)
	return EOPNOTSUPP
}

func Clearenv() {
	println("syscall.Clearenv not implemented")
}

func runtime_envs() []string

func Environ() []string {
	env := runtime_envs()
	envCopy := make([]string, len(env))
	copy(envCopy, env)
	return envCopy
}
