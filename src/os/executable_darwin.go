//go:build darwin

package os

// via runtime because we need argc/argv ptrs
func runtime_executable_path() string

func Executable() (string, error) {
	p := runtime_executable_path()
	if p != "" && p[0] == '/' {
		// absolute path
		return p, nil
	}
	cwd, err := Getwd()
	if err != nil {
		return "", err
	}
	return joinPath(cwd, p), nil
}
