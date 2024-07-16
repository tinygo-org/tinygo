package cm

// Tuple represents a [Component Model tuple] with 2 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple[T0, T1 any] struct {
	F0 T0
	F1 T1
}

// Tuple3 represents a [Component Model tuple] with 3 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple3[T0, T1, T2 any] struct {
	F0 T0
	F1 T1
	F2 T2
}

// Tuple4 represents a [Component Model tuple] with 4 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple4[T0, T1, T2, T3 any] struct {
	F0 T0
	F1 T1
	F2 T2
	F3 T3
}

// Tuple5 represents a [Component Model tuple] with 5 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple5[T0, T1, T2, T3, T4 any] struct {
	F0 T0
	F1 T1
	F2 T2
	F3 T3
	F4 T4
}

// Tuple6 represents a [Component Model tuple] with 6 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple6[T0, T1, T2, T3, T4, T5 any] struct {
	F0 T0
	F1 T1
	F2 T2
	F3 T3
	F4 T4
	F5 T5
}

// Tuple7 represents a [Component Model tuple] with 7 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple7[T0, T1, T2, T3, T4, T5, T6 any] struct {
	F0 T0
	F1 T1
	F2 T2
	F3 T3
	F4 T4
	F5 T5
	F6 T6
}

// Tuple8 represents a [Component Model tuple] with 8 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple8[T0, T1, T2, T3, T4, T5, T6, T7 any] struct {
	F0 T0
	F1 T1
	F2 T2
	F3 T3
	F4 T4
	F5 T5
	F6 T6
	F7 T7
}

// Tuple9 represents a [Component Model tuple] with 9 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple9[T0, T1, T2, T3, T4, T5, T6, T7, T8 any] struct {
	F0 T0
	F1 T1
	F2 T2
	F3 T3
	F4 T4
	F5 T5
	F6 T6
	F7 T7
	F8 T8
}

// Tuple10 represents a [Component Model tuple] with 10 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9 any] struct {
	F0 T0
	F1 T1
	F2 T2
	F3 T3
	F4 T4
	F5 T5
	F6 T6
	F7 T7
	F8 T8
	F9 T9
}

// Tuple11 represents a [Component Model tuple] with 11 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 any] struct {
	F0  T0
	F1  T1
	F2  T2
	F3  T3
	F4  T4
	F5  T5
	F6  T6
	F7  T7
	F8  T8
	F9  T9
	F10 T10
}

// Tuple12 represents a [Component Model tuple] with 12 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple12[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11 any] struct {
	F0  T0
	F1  T1
	F2  T2
	F3  T3
	F4  T4
	F5  T5
	F6  T6
	F7  T7
	F8  T8
	F9  T9
	F10 T10
	F11 T11
}

// Tuple13 represents a [Component Model tuple] with 13 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple13[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12 any] struct {
	F0  T0
	F1  T1
	F2  T2
	F3  T3
	F4  T4
	F5  T5
	F6  T6
	F7  T7
	F8  T8
	F9  T9
	F10 T10
	F11 T11
	F12 T12
}

// Tuple14 represents a [Component Model tuple] with 14 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple14[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13 any] struct {
	F0  T0
	F1  T1
	F2  T2
	F3  T3
	F4  T4
	F5  T5
	F6  T6
	F7  T7
	F8  T8
	F9  T9
	F10 T10
	F11 T11
	F12 T12
	F13 T13
}

// Tuple15 represents a [Component Model tuple] with 15 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple15[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14 any] struct {
	F0  T0
	F1  T1
	F2  T2
	F3  T3
	F4  T4
	F5  T5
	F6  T6
	F7  T7
	F8  T8
	F9  T9
	F10 T10
	F11 T11
	F12 T12
	F13 T13
	F14 T14
}

// Tuple16 represents a [Component Model tuple] with 16 fields.
//
// [Component Model tuple]: https://component-model.bytecodealliance.org/design/wit.html#tuples
type Tuple16[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15 any] struct {
	F0  T0
	F1  T1
	F2  T2
	F3  T3
	F4  T4
	F5  T5
	F6  T6
	F7  T7
	F8  T8
	F9  T9
	F10 T10
	F11 T11
	F12 T12
	F13 T13
	F14 T14
	F15 T15
}

// MaxTuple specifies the maximum number of fields in a Tuple* type, currently [Tuple16].
// See https://github.com/WebAssembly/component-model/issues/373 for more information.
const MaxTuple = 16
