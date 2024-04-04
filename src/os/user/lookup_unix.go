// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !baremetal && !js && !wasip1 && !windows
// +build !baremetal,!js,!wasip1,!windows

package user

import (
	"bufio"
	"os"
	"strings"
)

func lookupUser(username string) (*User, error) {
	f, err := os.Open("/etc/passwd")
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
