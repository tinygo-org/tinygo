.. _installation:

.. highlight:: none

Installation instructions
=========================

Requirements
------------

These are the base requirements and enough for most (desktop) use.

  * Go 1.11+
  * LLVM 7 (for example, from `apt.llvm.org <http://apt.llvm.org/>`_)

Linking a binary needs an installed C compiler (``cc``). At the moment it
expects GCC or a recent Clang.

ARM Cortex-M
~~~~~~~~~~~~

The Cortex-M family of microcontrollers is well supported, as it uses the stable
ARM LLVM backend (which is even used in the propietary C compiler from ARM).
Compiling to object code should be supported out of the box, but compiling the
final binary and flashing it needs some extra tools.

    * binutils (``arm-none-eabi-ld``, ``arm-none-eabi-objcopy``) for linking and
      for producing .hex files for flashing.
    * Clang 7 (``clang-7``) for building assembly files and the `compiler
      runtime library <https://compiler-rt.llvm.org/>`_ .
    * The flashing tool for the particular chip, like ``openocd`` or
      ``nrfjprog``.

AVR (Arduino)
~~~~~~~~~~~~~

The AVR backend has similar requirements as the `ARM Cortex-M`_ backend. It
needs the following tools:

    * binutils (``avr-objcopy``) for flashing.
    * GCC (``avr-gcc``) for linking object files.
    * libc (``avr-libc``), which is not installed on Debian as a dependency of
      ``avr-gcc``.
    * ``avrdude`` for flashing to an Arduino.

WebAssembly
~~~~~~~~~~~

The WebAssembly backend only needs a special linker from the LLVM project:

    * LLVM linker (``ld.lld-7``) for linking WebAssembly files together.


Installation
------------

First download the sources. This may take a while. ::

    go get -u github.com/aykevl/tinygo

If you get an error like this::

    /usr/local/go/pkg/tool/linux_amd64/link: running g++ failed: exit status 1
    /usr/bin/ld: error: cannot find -lLLVM-7
    cgo-gcc-prolog:58: error: undefined reference to 'LLVMVerifyFunction'
    cgo-gcc-prolog:80: error: undefined reference to 'LLVMVerifyModule'
    [...etc...]

Or like this::

    ../go-llvm/analysis.go:17:93: fatal error: llvm-c/Analysis.h: No such file or directory
     #include "llvm-c/Analysis.h" // If you are getting an error here read bindings/go/README.txt

It means something is wrong with your LLVM installation. Make sure LLVM 7 is
installed (Debian package ``llvm-7-dev``). If it still doesn't work, you can
try running::

    cd $GOPATH/github.com/aykevl/go-llvm
    make config

And retry::

    go install github.com/aykevl/tinygo

Usage
-----

TinyGo should now be installed. Test it by running a test program::

    tinygo run examples/test

Before anything can be built for a bare-metal target, you need to generate some
files first::

    make gen-device

This will generate register descriptions, interrupt vectors, and linker scripts
for various devices. Also, you may need to re-run this command after updates,
as some updates cause changes to the generated files.

Now you can run a blinky example. For the `PCA10040
<https://www.nordicsemi.com/eng/Products/Bluetooth-low-energy/nRF52-DK>`_
development board::

    tinygo flash -target=pca10040 examples/blinky2

Or for an `Arduino Uno <https://store.arduino.cc/arduino-uno-rev3>`_::

    tinygo flash -target=arduino examples/blinky1
