package inner

type Inner struct{}

func (Inner) value() int { return 1 }

type valuer interface{ value() int }

// Value calls the value method of package inner through an interface.
func Value(v any) int { return v.(valuer).value() }
