package value

type (
	Value[T any] interface {
		Get(Callback[T], Callback[T])
	}

	Callback[T any] func(T)

	Transform[S any, D any] func(S) D
)

func New[T any](v T) Value[T] {
	return &value[T]{
		v: v,
	}
}

type value[T any] struct {
	v T
}

func (v *value[T]) Get(fn1, fn2 Callback[T]) {
	// For example purposes.
	// Normally would be asynchronous callback.
	fn1(v.v)
	fn2(v.v)
}

func Map[S, D any](v Value[S], tx Transform[S, D]) Value[D] {
	return &mapper[S, D]{
		v:  v,
		tx: tx,
	}
}

type mapper[S, D any] struct {
	v  Value[S]
	tx Transform[S, D]
}

func (m *mapper[S, D]) Get(fn1, fn2 Callback[D]) {
	// two callbacks are passed to generate more than
	// one anonymous function symbol name.
	m.v.Get(func(v S) {
		// anonymous function inside of anonymous function.
		func() {
			fn1(m.tx(v))
		}()
	}, func(v S) {
		fn2(m.tx(v))
	})
}
