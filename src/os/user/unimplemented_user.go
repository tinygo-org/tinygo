package user

/*
   os/user stubbed functions

TODO: these should return unsupported / unimplemented errors

*/

import (
	"errors"
)

// Current returns the current user.
//
// The first call will cache the current user information.
// Subsequent calls will return the cached value and will not reflect
// changes to the current user.
func Current() (*User, error) {
	return nil, errors.New("user: Current not implemented")
}

// Lookup always returns an error.
func Lookup(username string) (*User, error) {
	return nil, errors.New("user: Lookup not implemented")
}

// LookupGroup always returns an error.
func LookupGroup(name string) (*Group, error) {
	return nil, errors.New("user: LookupGroup not implemented")
}
