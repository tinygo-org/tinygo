// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !baremetal && !js && !wasip1 && !windows
// +build !baremetal,!js,!wasip1,!windows

package user

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

const (
	userFile  = "/etc/passwd"
	groupFile = "/etc/group"
)

func lookupUser(username string) (*User, error) {
	f, err := os.Open(userFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// parse file format <username>:<password>:<uid>:<gid>:<gecos>:<home>:<shell>
	lines := bufio.NewScanner(f)
	for lines.Scan() {
		line := lines.Text()
		fragments := strings.Split(line, ":")

		if len(fragments) < 7 {
			continue
		}

		if fragments[0] == username {
			return &User{
				Uid:      fragments[2],
				Gid:      fragments[3],
				Username: fragments[0],
				Name:     fragments[4],
				HomeDir:  fragments[5],
			}, nil
		}
	}

	return nil, UnknownUserError(username)
}

func lookupUserId(uid string) (*User, error) {
	f, err := os.Open(userFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// parse file format <username>:<password>:<uid>:<gid>:<gecos>:<home>:<shell>
	lines := bufio.NewScanner(f)
	for lines.Scan() {
		line := lines.Text()
		fragments := strings.Split(line, ":")

		if len(fragments) < 7 {
			continue
		}

		if fragments[2] == uid {
			return &User{
				Uid:      fragments[2],
				Gid:      fragments[3],
				Username: fragments[0],
				Name:     fragments[4],
				HomeDir:  fragments[5],
			}, nil
		}
	}

	id, err := strconv.Atoi(uid)
	if err != nil {
		return nil, err
	}

	return nil, UnknownUserIdError(id)
}

func lookupGroupId(gid string) (*Group, error) {
	f, err := os.Open(groupFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// parse file format group_name:password:GID:user_list
	// group_name: the name of the group.
	// password: the (encrypted) group password. If this field is empty, no password is needed.
	// GID: the numeric group ID.
	// user_list: a list of the usernames that are members of this group, separated by commas.

	lines := bufio.NewScanner(f)
	for lines.Scan() {
		line := lines.Text()
		fragments := strings.Split(line, ":")

		if len(fragments) < 4 {
			continue
		}

		if fragments[2] == gid {
			return &Group{
				Gid:  fragments[2],
				Name: fragments[0],
			}, nil
		}
	}

	return nil, UnknownGroupIdError(gid)
}
