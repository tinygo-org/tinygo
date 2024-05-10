// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !baremetal && !js && !wasip1 && !windows
// +build !baremetal,!js,!wasip1,!windows

package os_test

import (
	. "os/user"
	"runtime"
	"testing"
)

// NOTE: This test requires some users to be present in the CI environment.
// If the test fails, it may be because the users are not present.
// We can guarantee that the root user is present on all systems.
// Currently we can also guarantee that the user with UID 1001 and GID 127 is present on all systems.
func TestUserLookup(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		t.Skip()
	}

	testCases := map[string]struct {
		user    User
		wantErr bool
	}{
		"root": {
			wantErr: false,
			user: User{
				Uid:      "0",
				Gid:      "0",
				Username: "root",
				Name:     "root",
				HomeDir:  "/root",
			},
		},
		"error": {
			wantErr: true,
			user: User{
				Uid:      "1000000",
				Gid:      "1000000",
				Username: "nonexistentuser",
				Name:     "nonexistentuser",
				HomeDir:  "/home/nonexistentuser",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			user, err := Lookup(tc.user.Username)
			if (err != nil && !tc.wantErr) || (err == nil && tc.wantErr) {
				t.Fatalf("Lookup(%q) = %v; want %v, got error %v", tc.user.Username, user, tc.user, err)
			}

			userId, err := LookupId(tc.user.Uid)
			if (err != nil && !tc.wantErr) || (err == nil && tc.wantErr) {
				t.Fatalf("LookupId(%q) = %v; want %v, got error %v", tc.user.Username, userId, tc.user, err)
			}

			group, err := LookupGroupId(tc.user.Gid)
			if (err != nil && !tc.wantErr) || (err == nil && tc.wantErr) {
				t.Fatalf("LookupGroupId(%q) = %v; want %v, got error %v", tc.user.Gid, group, tc.user.Gid, err)
			}
		})
	}
}
