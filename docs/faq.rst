.. faq:

Frequently Asked Questions
==========================


What is TinyGo exactly?
-----------------------

A new compiler and a new runtime implementation.

Specifically:

  * A new compiler using (mostly) the standard library to parse Go programs and
    using LLVM to optimize the code and generate machine code for the target
    architecture.

  * A new runtime library that implements some compiler intrinsics, like a
    memory allocator, a scheduler, and operations on strings. Also, some
    packages that are strongly connected to the runtime like the ``sync``
    package and the ``reflect`` package have been or will be re-implemented for
    use with this new compiler.


Why a new compiler?
-------------------

Why not modify the existing compiler to produce binaries for microcontrollers?

There are several reasons for this:

  * The standard Go compiler (``gc``) does not support instruction sets as used
    on microcontrollers:

      * The Thumb instruction set is unsupported, but it should be possible to
        add support for it as it already has an ARM backend.
      * The AVR instruction set (as used in the Arduino Uno) is unsupported and
        unlikely to be ever supported.

    Of course, it is possible to use ``gccgo``, but that has different problems
    (see below).

  * The runtime is really big. A standard 'hello world' on a desktop PC produces
    a binary of about 1MB, even when using the builtin ``println`` function and
    nothing else. All this overhead is due to the runtime. Of course, it may be
    possible to use a different runtime with the same compiler but that will be
    kind of painful as the exact ABI as used by the compiler has to be matched,
    limiting optimization opportunities (see below).

  * The compiler is optimized for speed, not for code size or memory
    consumption (which are usually far more important on MCUs). This results in
    design choices like allocating memory on every value → interface conversion
    while TinyGo sacrifices some performance for reduced GC pressure.

  * With the existing Go libraries for parsing Go code and the pretty awesome
    LLVM optimizer/backend it is relatively easy to get simple Go programs
    working with a very small binary size. Extra features can be added where
    needed in a pay-as-you-go manner similar to C++ avoiding their cost when
    unused. Most programs on microcontrollers are relatively small so a
    not-complete compiler is still useful.

  * The standard Go compilers do not allocate global variables as static data,
    but as zero-initialized data that is initialized during program startup.
    This is not a big deal on desktop computers but prevents allocating these
    values in flash on microcontrollers. Part of this is due to how the
    `language specification defines package initialization
    <https://golang.org/ref/spec#Package_initialization>`_, but this can be
    worked around to a large extent.

  * The standard Go compilers do a few special things for CGo calls. This is
    necessary because only Go code can use the (small) Go stack while C code
    will need a much bigger stack. A new compiler can avoid this limitation if
    it ensures stacks are big enough for C, greatly reducing the C ↔ Go calling
    overhead.

`At one point <https://github.com/aykevl/tinygo-gccgo>`_, a real Go compiler
had been used to produce binaries for various platforms, and the result was
painful enough to start writing a new compiler:

  * The ABI was fixed, so could not be optimized for speed. Also, the ABI
    didn't seem to be documented anywhere.

  * Working arount limitations in the ``go`` toolchain was rather burdensome
    and quite a big hack.

  * The binaries produced were quite bloated, for various reasons:

      * The Go calling convention places all arguments on the stack. Due to
        this, stack usage was really bad and code size was bigger than it
        needed to be.

      * Global initialization was very inefficient, see above.

      * There seemed to be no way to optimize across packages.


Why Go instead of Rust?
-----------------------

Rust is another "new" and safer language that is now made ready for embedded
processors. There is `a fairly active community around it
<https://rust-embedded.github.io/blog/>`_.

However, apart from personal language preference, Go has a few advantages:

  * Subjective, but in general Go is `easier to learn
    <https://matthias-endler.de/2017/go-vs-rust/>`_. Rust is in general far more
    complicated than Go, with difficult-to-grasp ownership rules, traits,
    generics, etc. Go prides itself on being a simple and slightly dumb
    language, sacrificing some expressiveness for readability.

  * Built-in support for concurrency with goroutines and channels that do not
    rely on a particular implementation threads. This avoids the need for a
    custom `RTOS-like framework <https://blog.japaric.io/rtfm-v2/>`_ or a
    `full-blown RTOS <https://github.com/rust-embedded/wg/issues/45>`_ with the
    associated API one has to learn. In Go, everything is handled by goroutines
    which are built into the language itself.

  * A batteries-included standard library that consists of loosely-coupled
    packages. Rust uses a monolithic standard library that is currently unusable
    on bare-metal, while the Go standard library is much more loosely coupled so
    is more likely to be (partially) supported. Also, non-standard packages in
    Go do not have to be marked with something like ``#![no_std]`` to be usable
    on bare metal. Note: most standard library packages cannot yet be compiled,
    but this situation will hopefully improve in the future.

At the same time, Rust has other advantages:

  * Unlike Go, Rust does not have a garbage collector by default and carefully
    written Rust code can avoid most or all uses of the heap. Go relies heavily
    on garbage collection and often implicitly allocates memory on the heap.

  * Rust has stronger memory-safety guarantees.

  * In general, Rust is more low-level and easier to support on a
    microcontroller. Of course, this doesn't mean one shouldn't try to run Go on
    a microcontroller, just that it is more difficult. When even dynamic
    languages like `Python <https://micropython.org/>`_, `Lua
    <https://nodemcu.readthedocs.io/en/master/>`_ and `JavaScript
    <https://www.espruino.com/>`_ can run on a microcontroller, then certainly
    Go can.


.. _faq-esp:

What about the ESP8266/ESP32?
-----------------------------

These chips use the rather obscure Xtensa instruction set. While a port of GCC
exists and Espressif provides precompiled GNU toolchains, there is no support
yet in LLVM (although there have been `multiple attempts
<http://lists.llvm.org/pipermail/llvm-dev/2018-July/124789.html>`_).

There are two ways these chips might be supported in the future, and both will
take a considerable amount of work:

  * The compiled LLVM IR can be converted into (ugly) C and then be compiled
    with a supported C compiler (like GCC for Xtensa). This has been `done
    before <https://github.com/JuliaComputing/llvm-cbe>`_ so should be doable.

  * One of the work-in-progress LLVM backends can be worked on to get it in a
    usable state. If this is finished, a true TinyGo port is possible.
