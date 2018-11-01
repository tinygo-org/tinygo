.. _internals:


Compiler internals
==================


Differences from ``go``
-----------------------

  * A whole program is compiled in a single step, without intermediate linking.
    This makes incremental development much slower for large programs but
    enables far more optimization opportunities. In the future, an option should
    be added for incremental compilation during edit-compile-test cycles.
  * Interfaces are always represented as a ``{typecode, value}`` pair. `Unlike
    Go <https://research.swtch.com/interfaces>`_, TinyGo will not precompute a
    list of function pointers for fast interface method calls. Instead, all
    interface method calls are looked up where they are used. This may sound
    expensive, but it avoids memory allocation at interface creation.
  * Global variables are computed during compilation whenever possible (unlike
    Go, which does not have the equivalent of a ``.data`` section). This is an
    important optimization for several reasons:
      * Startup time is reduced. This is nice, but not the main reason.
      * Initializing globals by copying the initial data from flash to RAM costs
        much less flash space as only the actual data needs to be stored,
        instead of all instructions to initialize these globals.
      * Data can often be statically allocated instead of dynamically allocated
        at startup.
      * Dead globals are trivially optimized away by LLVM.
      * Constant globals are trivially recognized by LLVM and marked
        ``constant``. This makes sure they can be stored in flash instead of
        RAM.
      * Global constants are useful for constant propagation and thus for dead
        code elimination (like an ``if`` that depends on a global variable).


Datatypes
---------

TinyGo uses a different representation for some data types than standard Go.

string
    A string is encoded as a ``{ptr, len}`` tuple. The type is actually defined
    in the runtime as ``runtime._string``, in `src/runtime/string.go
    <https://github.com/aykevl/tinygo/blob/master/src/runtime/string.go>`_. That
    file also contains some compiler intrinsics for dealing with strings and
    UTF-8.

slice
    A slice is encoded as a ``{ptr, len, cap}`` tuple. There is no runtime
    definition of it as slices are a generic type and the pointer type is
    different for each slice. That said, the bit layout is exactly the same for
    every slice and generic ``copy`` and ``append`` functions are implemented in
    `src/runtime/slice.go
    <https://github.com/aykevl/tinygo/blob/master/src/runtime/slice.go>`_.

array
    Arrays are simple: they are simply lowered to a LLVM array type.

complex
    Complex numbers are implemented in the most obvious way: as a vector of
    floating point numbers with length 2.

map
    The map type is a very complex type and is implemented as an (incomplete)
    hashmap. It is defined as ``runtime.hashmap`` in `src/runtime/hashmap.go
    <https://github.com/aykevl/tinygo/blob/master/src/runtime/hashmap.go>`_. As
    maps are reference types, they are lowered to a pointer to the
    aforementioned struct. See for example ``runtime.hashmapMake`` that is the
    compiler intrinsic to create a new hashmap.

interface
    An interface is a ``{typecode, value}`` tuple and is defined as
    ``runtime._interface`` in `src/runtime/interface.go
    <https://github.com/aykevl/tinygo/blob/master/src/runtime/interface.go>`_.
    The typecode is a small integer unique to the type of the value. See
    interface.go for a detailed description of how typeasserts and interface
    calls are implemented.

function pointer
    A function pointer has two representations: a literal function pointer and a
    fat function pointer in the form of ``{context, function pointer}``. Which
    representation is chosen depends on the AnalyseFunctionPointers pass in
    `ir/passes.go <https://github.com/aykevl/tinygo/blob/master/ir/passes.go>`_:
    it tries to use a raw function pointer but will use a fat function pointer
    if there is a closure or bound method somewhere in the program with the
    exact same signature.

goroutine
    A goroutine is a linked list of `LLVM coroutines
    <https://llvm.org/docs/Coroutines.html>`_. Every blocking call will create a
    new coroutine, pass the resulting coroutine to the scheduler, and will mark
    itself as waiting for a return. Once the called blocking function returns,
    it re-activates its parent coroutine. Non-blocking calls are normal calls,
    unaware of the fact that they're running on a particular goroutine. For
    details, see `src/runtime/scheduler.go
    <https://github.com/aykevl/tinygo/blob/master/src/runtime/scheduler.go>`_.

    This is rather expensive and should be optimized in the future. But the way
    it works now, a single stack can be used for all goroutines lowering memory
    consumption.


Calling convention
------------------

Go uses a stack-based calling convention and passes a pointer to the argument
list as the first argument in the function. There were/are `plans to switch to a
register-based calling convention <https://github.com/golang/go/issues/18597>`_
but they're now on hold.

.. highlight:: llvm

TinyGo, however, uses a register based calling convention. In fact it is
somewhat compatible with the C calling convention but with a few quirks:

  * Struct parameters are split into separate arguments, if the number of fields
    (after flattening recursively) is 3 or lower. This is similar to the `Swift
    calling convention
    <https://github.com/apple/swift/blob/master/docs/CallingConvention.rst#physical-conventions>`_.
    In the case of TinyGo, the size of each field does not matter, a field can
    even be an array. ::

      {i8*, i32}           -> i8*, i32
      {{i8*, i32}, i16}    -> i8*, i32, i16
      {{i64}}              -> i64
      {}                   ->
      {i8*, i32, i8, i8}   -> {i8*, i32, i8, i8}
      {{i8*, i32, i8}, i8} -> {i8*, i32, i8, i8}

    Note that all native Go data types that are lowered to aggregate types in
    LLVM are expanded this way: ``string``, slices, interfaces, and fat function
    pointers. This avoids some overhead in the C calling convention and makes
    the work of the LLVM optimizers easier.

  * Some functions have an extra context parameter appended at the end of the
    argument list. This only happens when both of these conditions hold:

      * The address of the function is taken, for example when passing the
        function as function pointer to another function or storing it in a
        global variable.

      * This function or another function somewhere in the compiled code has the
        exact same signature and is used in a closure or bound method. Signature
        matching is very strict: it is based on Go types including named types
        and return types, although parameter names or the function name itself
        are not included in the match.

    Whether a function needs this is determined by `FunctionNeedsContext
    <https://godoc.org/github.com/aykevl/tinygo/ir#Program.FunctionNeedsContext>`_,
    which bases itself on analysis done by AnalyseFunctionPointers.

  * Blocking functions have a coroutine pointer prepended to the argument list,
    see `src/runtime/scheduler.go
    <https://github.com/aykevl/tinygo/blob/master/src/runtime/scheduler.go>`_
    for details. Whether a function is blocking is determined by the
    AnalyseBlockingRecursive pass.

This calling convention may change in the future. Changes will be documented
here. However, even though it may change, it is expected that function
signatures that only contain integers and pointers will remain stable.


Pipeline
--------

Like most compilers, TinyGo is a compiler built as a pipeline of
transformations, that each translate an input to a simpler output version (also
called lowering). However, most of these part are not in TinyGo itself. The
frontend is mostly implemented by external Go libraries, and most optimizations
and code generation is implemented by LLVM.

This is roughly the pipeline for TinyGo:

  * Lexing, parsing, typechecking and `AST
    <https://en.wikipedia.org/wiki/Abstract_syntax_tree>`_ building is done by
    packages in the `standard library <https://godoc.org/go>`_ and in the
    `golang.org/x/tools/go library <https://godoc.org/golang.org/x/tools/go>`_.
  * `SSA <https://en.wikipedia.org/wiki/Static_single_assignment_form>`_
    construction (a very important step) is done by the
    `golang.org/x/tools/go/ssa <https://godoc.org/golang.org/x/tools/go/ssa>`_
    package.
  * This SSA form is then analyzed by the `ir package
    <https://godoc.org/github.com/aykevl/tinygo/ir>`_ to learn all kinds of
    things about the code that help the optimizer.
  * The Go SSA is then transformed into LLVM IR by the `compiler package
    <https://godoc.org/github.com/aykevl/tinygo/compiler>`_. Both forms are SSA,
    but because Go SSA is higher level and contains Go-specific constructs (like
    interfaces and goroutines) this is non-trivial. However, the vast majority
    of the work is simply lowering the available Go SSA into LLVM IR, possibly
    calling some runtime library intrinsics in the process (for example,
    operations on maps).
  * This LLVM IR is then optimized by the LLVM optimizer, which has a large
    array of standard `optimization passes
    <https://llvm.org/docs/Passes.html>`_. Currently, the standard optimization
    pipeline is used as is also be used by Clang, but a pipeline better tuned
    for TinyGo might be used in the future.
  * After all optimizations have run, a few fixups are needed for AVR for
    globals. This is implemented by the compiler package.
  * Finally, the resulting machine code is emitted by LLVM to an object file, to
    be linked by an architecture-specific linker in a later step.

After this whole list of compiler phases, the Go source has been transformed
into object code. It can then be emitted directly to a file (for linking in a
different build system), or it can be linked directly or even be flashed to a
target by TinyGo (using external tools under the hood).
