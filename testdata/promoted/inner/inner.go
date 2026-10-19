package inner

type Inner struct{}

func (Inner) value() int { return 1 }

func (*Inner) ptr() int { return 1 }

type valuer interface{ value() int }

type ptrer interface{ ptr() int }

// Value calls the value method of package inner through an interface.
func Value(v any) int { return v.(valuer).value() }

// Ptr calls the pointer method of package inner through an interface.
func Ptr(v any) int { return v.(ptrer).ptr() }
