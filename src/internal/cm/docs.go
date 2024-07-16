// Package cm contains types and functions for interfacing with the WebAssembly Component Model.
//
// The types in this package (such as [List], [Option], [Result], and [Variant]) are designed to match the memory layout
// of [Component Model] types as specified in the [Canonical ABI].
//
// [Component Model]: https://component-model.bytecodealliance.org/introduction.html
// [Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
package cm
