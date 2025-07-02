package cm

// CaseUnmarshaler returns an function that can unmarshal text into
// [variant] or [enum] case T.
//
// [enum]: https://component-model.bytecodealliance.org/design/wit.html#enums
// [variant]: https://component-model.bytecodealliance.org/design/wit.html#variants
func CaseUnmarshaler[T ~uint8 | ~uint16 | ~uint32](cases []string) func(v *T, text []byte) error {
	if len(cases) <= linearScanThreshold {
		return func(v *T, text []byte) error {
			if len(text) == 0 {
				return &emptyTextError{}
			}
			s := string(text)
			for i := 0; i < len(cases); i++ {
				if cases[i] == s {
					*v = T(i)
					return nil
				}
			}
			return &noMatchingCaseError{}
		}
	}

	m := make(map[string]T, len(cases))
	for i, v := range cases {
		m[v] = T(i)
	}

	return func(v *T, text []byte) error {
		if len(text) == 0 {
			return &emptyTextError{}
		}
		c, ok := m[string(text)]
		if !ok {
			return &noMatchingCaseError{}
		}
		*v = c
		return nil
	}
}

const linearScanThreshold = 16

type emptyTextError struct{}

func (*emptyTextError) Error() string { return "empty text" }

type noMatchingCaseError struct{}

func (*noMatchingCaseError) Error() string { return "no matching case" }
