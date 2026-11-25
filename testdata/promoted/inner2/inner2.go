package inner2

type Inner2 struct{}

func (Inner2) value() int { return 3 }

type valuer interface{ value() int }

func Value(v any) int { return v.(valuer).value() }
