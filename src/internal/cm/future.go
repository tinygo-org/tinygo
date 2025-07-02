package cm

// Future represents the Component Model [future] type.
// A future is a special case of stream. In non-error cases,
// a future delivers exactly one value before being automatically closed.
//
// [future]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/Explainer.md#asynchronous-value-types
type Future[T any] struct {
	_ HostLayout
	future[T]
}

type future[T any] uint32

// TODO: implement methods on type future
