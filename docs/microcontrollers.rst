.. _microcontrollers:

.. highlight:: go


Go on microcontrollers
======================

TinyGo was designed to run on microcontrollers, but the Go language wasn't.
This means there are a few challenges to writing Go code for microcontrollers.

Microcontrollers have very little RAM and execute code directly from flash.
Also, constant globals are generally put in flash whenever possible. The Go
language itself heavily relies on garbage collection so care must be taken to
avoid dynamic memory allocation.


Heap allocation
---------------

Many operations in Go rely on heap allocation. Some of these heap allocations
are optimized away, but not all of them. Also, TinyGo does not yet contain a
garbage collector so heap allocation must be avoided whenever possible outside
of initialization code.

These operations currently do heap allocations:

  * Taking the pointer of a local variable. This will result in a heap
    allocation, unless the compiler can see the resulting pointer never
    escapes. This causes a heap allocation::

        var global *int

        func foo() {
            i := 3
            global = &i
        }

    This does not cause a heap allocation::

        func foo() {
            i := 3
            bar(&i)
        }

        func bar(i *int) {
            println(*i)
        }

  * Converting between ``string`` and ``[]byte``. In general, this causes a
    heap allocation because one is constant while the other is not: for
    example, a ``[]byte`` is not allowed to write to the underlying buffer of a
    ``string``. However, there is an optimization that avoids a heap allocation
    when converting a string to a ``[]byte`` when the compiler can see the
    slice is never written to. For example, this ``WriteString`` function does
    not cause a heap allocation::

        func WriteString(s string) {
            Write([]byte(s))
        }

        func Write(buf []byte) {
            for _, c := range buf {
                WriteByte(c)
            }
        }

  * Converting a ``byte`` or ``rune`` into a ``string``. This operation is
    actually a conversion from a Unicode code point into a single-character
    string so is similar to the previous point.

  * Concatenating strings, unless one of them is zero length.

  * Creating an interface with a value larger than a pointer. Interfaces in Go
    are not a zero-cost abstraction and should be used carefully on
    microcontrollers.

  * Closures where the collection of shared variables between the closure and
    the main function is larger than a pointer.

  * Creating and modifying maps. Maps have *very* little support at the moment
    and should not yet be used. They exist mostly for compatibility with some
    standard library packages.

  * Starting goroutines. There is limited support for goroutines and currently
    they are not at all efficient. Also, there is no support for channels yet
    so their usefulness is limited.


The ``volatile`` keyword
------------------------

Go does not have the ``volatile`` keyword like C/C++. This keyword is
unnecessary in most desktop use cases but is required for memory mapped I/O on
microcontrollers and interrupt handlers. As a workaround, any variable of a
type annotated with the ``//go:volatile`` pragma will be marked volatile. For
example::

    //go:volatile
    type volatileBool bool

    var isrFlag volatileBool

This is a workaround for a limitation in the Go language and should at some
point be replaced with something else.


Inline assembly
---------------

The device-specific packages like ``device/avr`` and ``device/arm`` provide
``Asm`` functions which you can use to write inline assembly::

    arm.Asm("wfi")

You can also pass parameters to the inline assembly::

    var result int32
    arm.AsmFull(`
        lsls  {value}, #1
        str   {value}, {result}
    `, map[string]interface{}{
        "value":  42,
        "result": &result,
    })
    println("result:", result)

In general, types are autodetected. That is, integer types are passed as raw
registers and pointer types are passed as memory locations. This means you can't
easily do pointer arithmetic. To do that, convert a pointer value to a
``uintptr``.

Inline assembly support is expected to change in the future and may change in a
backwards-incompatible manner.


Harvard architectures (AVR)
---------------------------

The AVR architecture is a modified Harvard architecture, which means that flash
and RAM live in different address spaces. In practice, this means that any
given pointer may either point to RAM or flash, but this is not visible from
the pointer itself.

To get TinyGo to work on the Arduino, which uses the AVR architecutre, all
global variables (which include string constants!) are marked non-constant and
thus are stored in RAM and all pointer dereferences assume that pointers point
to RAM. At some point this should be optimized so that obviously constant data
is kept in read-only memory but this optimization has not yet been implemented.
