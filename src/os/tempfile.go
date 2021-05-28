package os

func CreateTemp(dir, pattern string) (*File, error) {
	return nil, &PathError{"createtemp", pattern, ErrNotImplemented}
}
