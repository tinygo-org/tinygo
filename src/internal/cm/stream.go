package cm

// Stream represents the Component Model [stream] type.
// A stream is a special case of stream. In non-error cases,
// a stream delivers exactly one value before being automatically closed.
//
// [stream]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/Explainer.md#asynchronous-value-types
type Stream[T any] struct {
	_ HostLayout
	stream[T]
}

type stream[T any] uint32

// TODO: implement methods on type stream
