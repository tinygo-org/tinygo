// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package user

import (
	"errors"
	"strconv"
)

// User represents a user account.
type User struct {
	// Uid is the user ID.
	// On POSIX systems, this is a decimal number representing the uid.
	// On Windows, this is a security identifier (SID) in a string format.
	// On Plan 9, this is the contents of /dev/user.
	Uid string
	// Gid is the primary group ID.
	// On POSIX systems, this is a decimal number representing the gid.
	// On Windows, this is a SID in a string format.
	// On Plan 9, this is the contents of /dev/user.
	Gid string
	// Username is the login name.
	Username string
	// Name is the user's real or display name.
	// It might be blank.
	// On POSIX systems, this is the first (or only) entry in the GECOS field
	// list.
	// On Windows, this is the user's display name.
	// On Plan 9, this is the contents of /dev/user.
	Name string
	// HomeDir is the path to the user's home directory (if they have one).
	HomeDir string
}

// UnknownUserError is returned by Lookup when a user cannot be found.
type UnknownUserError string

func (e UnknownUserError) Error() string {
	return "user: unknown user " + string(e)
}

// Group represents a grouping of users.
//
// On POSIX systems Gid contains a decimal number representing the group ID.
type Group struct {
	Gid  string // group ID
	Name string // group name
}

// UnknownGroupIdError is returned by [LookupGroupId] when
// a group cannot be found.
type UnknownGroupIdError string

func (e UnknownGroupIdError) Error() string {
	return "group: unknown groupid " + string(e)
}

// UnknownUserIdError is returned by [LookupId] when a user cannot be found.
type UnknownUserIdError int

func (e UnknownUserIdError) Error() string {
	return "user: unknown userid " + strconv.Itoa(int(e))
}

// Current returns the current user.
//
// The first call will cache the current user information.
// Subsequent calls will return the cached value and will not reflect
// changes to the current user.
// TODO: implement syscall.Getuid() and syscall.Getgid() to get the current user.
func Current() (*User, error) {
	return nil, errors.New("user: Current not implemented")
}

// Lookup looks up a user by username.
//
// If the user cannot be found, the returned error is of type UnknownUserError.
// NOTE: This implementation does not support caching as the golang implementation does.
func Lookup(username string) (*User, error) {
	return lookupUser(username)
}

// LookupId looks up a user by userid. If the user cannot be found, the
// returned error is of type UnknownUserIdError.
func LookupId(uid string) (*User, error) {
	if u, err := Current(); err == nil && u.Uid == uid {
		return u, err
	}
	return lookupUserId(uid)
}

// LookupGroupId looks up a group by groupid. If the group cannot be found, the
// returned error is of type [UnknownGroupIdError].
func LookupGroupId(gid string) (*Group, error) {
	return lookupGroupId(gid)
}
