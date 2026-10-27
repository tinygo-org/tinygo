package main

import "errors"

var (
	ErrTimeout  = errors.New("timeout")
	ErrNotFound = errors.New("device not found")
	ErrBusy     = errors.New("bus busy")
)

func check(code int) error {
	switch code {
	case 1:
		return ErrTimeout
	case 2:
		return ErrNotFound
	case 3:
		return ErrBusy
	}
	return nil
}

func main() {
	for code := range 4 {
		if err := check(code); err != nil {
			println(err.Error())
		}
	}
}
