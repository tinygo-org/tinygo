//go:build darwin

package os

// via runtime because we need argc/argv ptrs
func runtime_executable_path() string

func Executable() (string, error) {
	return runtime_executable_path(), nil
}
