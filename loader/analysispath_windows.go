package loader

import (
	"os"
	"strings"

	"golang.org/x/sys/windows"
)

func analysisPath(name string) (string, error) {
	file, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()
	buffer := make([]uint16, 260)
	for {
		n, err := windows.GetFinalPathNameByHandle(windows.Handle(file.Fd()), &buffer[0], uint32(len(buffer)), 0)
		if err != nil {
			return "", &os.PathError{Op: "GetFinalPathNameByHandle", Path: name, Err: err}
		}
		if n >= uint32(len(buffer)) {
			buffer = make([]uint16, n+1)
			continue
		}
		path := windows.UTF16ToString(buffer[:n])
		if strings.HasPrefix(path, `\\?\UNC\`) {
			return `\\` + strings.TrimPrefix(path, `\\?\UNC\`), nil
		}
		return strings.TrimPrefix(path, `\\?\`), nil
	}
}
