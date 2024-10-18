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

func TestWantGoVersion(t *testing.T) {
	tests := []struct {
		v     string
		major int
		minor int
		want  bool
	}{
		{"", 0, 0, false},
		{"go", 0, 0, false},
		{"go1", 0, 0, false},
		{"go.0", 0, 0, false},
		{"go1.0", 1, 0, true},
		{"go1.1", 1, 1, true},
		{"go1.23", 1, 23, true},
		{"go1.23.5", 1, 23, true},
		{"go1.23.5-rc6", 1, 23, true},
		{"go2.0", 1, 23, true},
		{"go2.0", 2, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.v, func(t *testing.T) {
			got := WantGoVersion(tt.v, tt.major, tt.minor)
			if got != tt.want {
				t.Errorf("WantGoVersion(%q, %d, %d): expected %t; got %t",
					tt.v, tt.major, tt.minor, tt.want, got)
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
		// {"go1.23.2", "go1.23.10", -1}, // FIXME: parse patch number
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
