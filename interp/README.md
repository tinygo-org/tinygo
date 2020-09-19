# Partial evaluation of initialization code in Go

For several reasons related to code size and memory consumption (see below), it
is best to try to evaluate as much initialization code at compile time as
possible and only run unknown expressions (e.g. external calls) at runtime. This
is in practice a partial evaluator of the `runtime.initAll` function, which
calls each package initializer.

It works by directly interpreting LLVM IR:

  * Almost all operations work directly on constants, and are implemented using
    the llvm.Const* set of functions that are evaluated directly.
  * External function calls and some other operations (inline assembly, volatile
    load, volatile store) are seen as having limited side effects. Limited in
    the sense that it is known at compile time which globals it affects, which
    then are marked 'dirty' (meaning, further operations on it must be done at
    runtime). These operations are emitted directly in the `runtime.initAll`
    function. Return values are also considered 'dirty'.
  * Such 'dirty' objects and local values must be executed at runtime instead of
    at compile time. This dirtyness propagates further through the IR, for
    example storing a dirty local value to a global also makes the global dirty,
    meaning that the global may not be read or written at compile time as it's
    contents at that point during interpretation is unknown.
  * There are some heuristics in place to avoid doing too much with dirty
    values. For example, a branch based on a dirty local marks the whole
    function itself as having side effect (as if it is an external function).
    However, all globals it touches are still taken into account and when a call
    is inserted in `runtime.initAll`, all globals it references are also marked
    dirty.
  * Heap allocation (`runtime.alloc`) is emulated by creating new objects. The
    value in the allocation is the initializer of the global, the zero value is
    the zero initializer.
  * Stack allocation (`alloca`) is often emulated using a fake alloca object,
    until the address of the alloca is taken in which case it is also created as
    a real `alloca` in `runtime.initAll` and marked dirty. This may be necessary
    when calling an external function with the given alloca as paramter.

## Why is this necessary?

A partial evaluator is hard to get right, so why go through all the trouble of
writing one?

The main reason is that the previous attempt wasn't complete and wasn't sound.
It simply tried to evaluate Go SSA directly, which was good but more difficult
than necessary. An IR based interpreter needs to understand fewer instructions
as the LLVM IR simply has less (complex) instructions than Go SSA. Also, LLVM
provides some useful tools like easily getting all uses of a function or global,
which Go SSA does not provide.

But why is it necessary at all? The answer is that globals with initializers are
much easier to optimize by LLVM than initialization code. Also, there are a few
other benefits:

* Dead globals are trivial to optimize away.
* Constant globals are easier to detect. Remember that Go does not have global
    constants in the same sense as that C has them. Constants are useful because
    they can be propagated and provide some opportunities for other
    optimizations (like dead code elimination when branching on the contents of
    a global).
* Constants are much more efficent on microcontrollers, as they can be
    allocated in flash instead of RAM.

For more details, see [this section of the
documentation](https://tinygo.org/compiler-internals/differences-from-go/).
