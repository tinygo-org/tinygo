// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	"errors"
	. "os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateTemp(t *testing.T) {
	nonexistentDir := filepath.Join("_also_not_exists_", "_not_exists_")
	f, err := CreateTemp(nonexistentDir, "foo")
	if f != nil || err == nil {
		t.Errorf("CreateTemp(%q, `foo`) = %v, %v", nonexistentDir, f, err)
	}
}

func TestCreateTempPattern(t *testing.T) {
	tests := []struct{ pattern, prefix, suffix string }{
		{"tempfile_test", "tempfile_test", ""},
		{"tempfile_test*", "tempfile_test", ""},
		{"tempfile_test*xyz", "tempfile_test", "xyz"},
	}
	for _, test := range tests {
		f, err := CreateTemp("", test.pattern)
		if err != nil {
			t.Errorf("CreateTemp(..., %q) error: %v", test.pattern, err)
			continue
		}
		defer Remove(f.Name())
		base := filepath.Base(f.Name())
		f.Close()
		if !(strings.HasPrefix(base, test.prefix) && strings.HasSuffix(base, test.suffix)) {
			t.Errorf("CreateTemp pattern %q created bad name %q; want prefix %q & suffix %q",
				test.pattern, base, test.prefix, test.suffix)
		}
	}
}

func TestCreateTempBadPattern(t *testing.T) {
	tmpDir := TempDir()

	const sep = string(PathSeparator)
	tests := []struct {
		pattern string
		wantErr bool
	}{
		{"ioutil*test", false},
		{"tempfile_test*foo", false},
		{"tempfile_test" + sep + "foo", true},
		{"tempfile_test*" + sep + "foo", true},
		{"tempfile_test" + sep + "*foo", true},
		{sep + "tempfile_test" + sep + "*foo", true},
		{"tempfile_test*foo" + sep, true},
	}
	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			tmpfile, err := CreateTemp(tmpDir, tt.pattern)
			if tmpfile != nil {
				defer tmpfile.Close()
			}
			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateTemp(..., %#q) succeeded, expected error", tt.pattern)
				}
				if !errors.Is(err, ErrPatternHasSeparator) {
					t.Logf("TODO: CreateTemp(..., %#q): %v, expected ErrPatternHasSeparator", tt.pattern, err)
				}
			} else if err != nil {
				t.Errorf("CreateTemp(..., %#q): %v", tt.pattern, err)
			}
		})
	}
}
