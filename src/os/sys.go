package os

func Hostname() (name string, err error) {
	return hostname()
}
