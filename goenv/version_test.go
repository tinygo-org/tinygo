package goenv

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		v       string
		major   int
		minor   int
		wantErr bool
	}{
		{"", 0, 0, true},
		{"go", 0, 0, true},
		{"go1", 0, 0, true},
		{"go.0", 0, 0, true},
		{"go1.0", 1, 0, false},
		{"go1.1", 1, 1, false},
		{"go1.23", 1, 23, false},
		{"go1.23.5", 1, 23, false},
		{"go1.23.5-rc6", 1, 23, false},
		{"go2.0", 2, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.v, func(t *testing.T) {
			major, minor, err := Parse(tt.v)
			if err == nil && tt.wantErr {
				t.Errorf("Parse(%q): expected err != nil", tt.v)
			}
			if err != nil && !tt.wantErr {
				t.Errorf("Parse(%q): expected err == nil", tt.v)
			}
			if major != tt.major || minor != tt.minor {
				t.Errorf("Parse(%q): expected %d, %d, nil; got %d, %d, %v",
					tt.v, tt.major, tt.minor, major, minor, err)
			}
		})
	}
}
