// Package godebug is a simplified version of the upstream internal/godebug.
// TinyGo only sees GODEBUG changes made with os.Setenv.
package godebug

import _ "unsafe" // go:linkname

// A Setting is a single setting in the $GODEBUG environment variable.
type Setting struct {
	name string
}

// The current value of the $GODEBUG environment variable.
var env string

// New returns a new Setting for the $GODEBUG setting with the given name.
// A name with a # prefix is an undocumented setting.
func New(name string) *Setting {
	return &Setting{name: name}
}

// Name returns the name of the setting.
func (s *Setting) Name() string {
	if s.name != "" && s.name[0] == '#' {
		return s.name[1:]
	}
	return s.name
}

// Undocumented reports whether this is an undocumented setting.
func (s *Setting) Undocumented() bool {
	return s.name != "" && s.name[0] == '#'
}

// String returns a printable form for the setting: name=value.
func (s *Setting) String() string {
	return s.Name() + "=" + s.Value()
}

// IncNonDefault does nothing because TinyGo has no runtime/metrics support.
func (s *Setting) IncNonDefault() {}

// Value returns the current value for the GODEBUG setting s.
func (s *Setting) Value() string {
	return lookup(env, s.Name())
}

// lookup returns the value of name in s. Later settings win and a #pattern
// suffix for bisect is ignored.
func lookup(s, name string) string {
	end := len(s)
	eq := -1
	for i := end - 1; i >= -1; i-- {
		if i == -1 || s[i] == ',' {
			if eq >= 0 && s[i+1:eq] == name {
				arg := s[eq+1 : end]
				for j := 0; j < len(arg); j++ {
					if arg[j] == '#' {
						return arg[:j]
					}
				}
				return arg
			}
			eq = -1
			end = i
		} else if s[i] == '=' {
			eq = i
		}
	}
	return ""
}

//go:linkname setUpdate
func setUpdate(update func(string, string))

func init() {
	setUpdate(update)
}

// update is called by the runtime when the GODEBUG environment variable
// changes.
func update(_, e string) {
	env = e
}
