//go:build linux && !baremetal && !tinygo.wasm && !nintendoswitch

package os

import "syscall"

// Hostname tries uname and then procfs, as Go's linux implementation does.
func Hostname() (name string, err error) {
	var un syscall.Utsname
	if err := syscall.Uname(&un); err == nil {
		b := make([]byte, 0, len(un.Nodename))
		for _, c := range un.Nodename {
			if c == 0 {
				break
			}
			b = append(b, byte(c))
		}
		if len(b) > 0 {
			return string(b), nil
		}
	}
	data, err := ReadFile("/proc/sys/kernel/hostname")
	if err != nil {
		return "", NewSyscallError("hostname", err)
	}
	for len(data) > 0 && (data[len(data)-1] == '\n' || data[len(data)-1] == 0) {
		data = data[:len(data)-1]
	}
	return string(data), nil
}
