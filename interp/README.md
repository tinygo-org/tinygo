# Partial evaluation of initialization code in Go

For several reasons related to code size and memory consumption (see below), it
is best to try to evaluate as much initialization code at compile time as
possible and only run unknown expressions (e.g. external calls) at runtime. This
is in practice a partial evaluator of the `runtime.initAll` function, which
calls each package initializer.

This package is a rewrite of a previous partial evaluator that worked
directly on LLVM IR and used the module and LLVM constants as intermediate
values. This newer version instead uses a mostly Go intermediate form. It
compiles functions and extracts relevant data first (compiler.go), then
executes those functions (interpreter.go) in a memory space that can be
rolled back per function (memory.go). This means that it is not necessary to
scan functions to see whether they can be run at compile time, which was very
error prone. Instead it just tries to execute everything and if it hits
something it cannot interpret (such as a store to memory-mapped I/O) it rolls
back the execution of that function and runs the function at runtime instead.
All in all, this design provides several benefits:

  * Much better error handling. By being able to revert to runtime execution
    without the need for scanning functions, this version is able to
    automatically work around many bugs in the previous implementation.
  * More correct memory model. This is not inherent to the new design, but the
    new design also made the memory model easier to reason about.
  * Faster execution of initialization code. While it is not much faster for
    normal interpretation (maybe 25% or so) due to the compilation overhead,
    it should be a whole lot faster for loops as it doesn't have to call into
    LLVM (via CGo) for every operation.

As mentioned, this partial evaluator comes in three parts: a compiler, an
interpreter, and a memory manager.

## Compiler

The main task of the compiler is that it extracts all necessary data from
every instruction in a function so that when this instruction is interpreted,
no additional CGo calls are necessary. This is not currently done for all
instructions (`runtime.alloc` is a notable exception), but at least it does
so for the vast majority of instructions.

## Interpreter

The interpreter runs an instruction just like it would if it were executed
'for real'. The vast majority of instructions can be executed at compile
time. As indicated above, some instructions need to be executed at runtime
instead.

## Memory

Memory is represented as objects (the `object` type) that contains data that
will eventually be stored in a global and values (the `value` interface) that
can be worked with while running the interpreter. Values therefore are only
used locally and are always passed by value (just like most LLVM constants)
while objects represent the backing storage (like LLVM globals). Some values
are pointer values, and point to an object.

Importantly, this partial evaluator can roll back the execution of a
function. This is implemented by creating a new memory view per function
activation, which makes sure that any change to a global (such as a store
instruction) is stored in the memory view. It creates a copy of the object
and stores that in the memory view to be modified. Once the function has
executed successfully, all these modified objects are then copied into the
parent function, up to the root function invocation which (on successful
execution) writes the values back into the LLVM module. This way, function
invocations can be rolled back without leaving a trace.

Pointer values point to memory objects, but not to a particular memory
object. Every memory object is given an index, and pointers use that index to
look up the current active object for the pointer to load from or to copy
when storing to it.

Rolling back a function should roll back everything, including the few
instructions emitted at runtime. This is done by treating instructions much
like memory objects and removing the created instructions when necessary.

## Why is this necessary?

A partial evaluator is hard to get right, so why go through all the trouble of
writing one?

The answer is that globals with initializers are much easier to optimize by
LLVM than initialization code. Also, there are a few other benefits:

  * Dead globals are trivial to optimize away.
  * Constant globals are easier to detect. Remember that Go does not have global
    constants in the same sense as that C has them. Constants are useful because
    they can be propagated and provide some opportunities for other
    optimizations (like dead code elimination when branching on the contents of
    a global).
  * Constants are much more efficient on microcontrollers, as they can be
    allocated in flash instead of RAM.

The Go SSA package does not create constant initializers for globals.
Instead, it emits initialization functions, so if you write the following:

```go
var foo = []byte{1, 2, 3, 4}
```

It would generate something like this:

```go
var foo []byte

func init() {
    foo = make([]byte, 4)
    foo[0] = 1
    foo[1] = 2
    foo[2] = 3
    foo[3] = 4
}
```

This is of course hugely wasteful, it's much better to create `foo` as a
global array instead of initializing it at runtime.

For more details, see [this section of the
documentation](https://tinygo.org/compiler-internals/differences-from-go/).
