package goenv

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		v       string
		major   int
		minor   int
		patch   int
		wantErr bool
	}{
		{"", 0, 0, 0, true},
		{"go", 0, 0, 0, true},
		{"go1", 0, 0, 0, true},
		{"go.0", 0, 0, 0, true},
		{"go1.0", 1, 0, 0, false},
		{"go1.1", 1, 1, 0, false},
		{"go1.23", 1, 23, 0, false},
		{"go1.23.5", 1, 23, 5, false},
		{"go1.23.5-rc6", 1, 23, 5, false},
		{"go2.0", 2, 0, 0, false},
		{"go2.0.15", 2, 0, 15, false},
	}
	for _, tt := range tests {
		t.Run(tt.v, func(t *testing.T) {
			major, minor, patch, err := Parse(tt.v)
			if err == nil && tt.wantErr {
				t.Errorf("Parse(%q): expected err != nil", tt.v)
			}
			if err != nil && !tt.wantErr {
				t.Errorf("Parse(%q): expected err == nil", tt.v)
			}
			if major != tt.major || minor != tt.minor || patch != tt.patch {
				t.Errorf("Parse(%q): expected %d, %d, %d, nil; got %d, %d, %d, %v",
					tt.v, tt.major, tt.minor, tt.patch, major, minor, patch, err)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want int
	}{
		{"", "", 0},
		{"go0", "go0", 0},
		{"go0", "go1", -1},
		{"go1", "go0", 1},
		{"go1", "go2", -1},
		{"go2", "go1", 1},
		{"go1.1", "go1.2", -1},
		{"go1.2", "go1.1", 1},
		{"go1.1.0", "go1.2.0", -1},
		{"go1.2.0", "go1.1.0", 1},
		{"go1.2.0", "go2.3.0", -1},
		{"go1.23.2", "go1.23.10", -1},
		{"go0.1.22", "go1.23.101", -1},
	}
	for _, tt := range tests {
		t.Run(tt.a+" "+tt.b, func(t *testing.T) {
			got := Compare(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Compare(%q, %q): expected %d; got %d",
					tt.a, tt.b, tt.want, got)
			}
		})
	}
}
