.. installation:

.. highlight:: none

Installation instructions
=========================

Requirements
------------

These are the base requirements and enough for most (desktop) use.

  * Go 1.11+
  * LLVM dependencies, see the Software section in the
    [LLVM build guide](https://llvm.org/docs/GettingStarted.html#software)

Linking a binary needs an installed C compiler (``cc``). At the moment it
expects GCC or a recent Clang.

ARM Cortex-M
~~~~~~~~~~~~

The Cortex-M family of microcontrollers is well supported, as it uses the stable
ARM LLVM backend (which is even used in the propietary C compiler from ARM).
Compiling to object code should be supported out of the box, but compiling the
final binary and flashing it needs some extra tools.

    * binutils (``arm-none-eabi-objcopy``) for producing .hex files for
      flashing.
    * GCC (``arm-none-eabi-gcc``) for linking object files.
    * The flashing tool for the particular chip, like ``openocd`` or
      ``nrfjprog``.

AVR (Arduino)
~~~~~~~~~~~~~

The AVR backend has similar requirements as the `ARM Cortex-M`_ backend. It
needs the following tools:

    * binutils (``avr-objcopy``) for flashing.
    * GCC (``avr-gcc``) for linking object files.
    * ``avrdude`` for flashing to an Arduino.


Installation
------------

First download the sources. This may take a few minutes. ::

    go get -u github.com/aykevl/tinygo

You'll get an error like the following, this is expected::

    src/github.com/aykevl/llvm/bindings/go/llvm/analysis.go:17:10: fatal error: 'llvm-c/Analysis.h' file not found
    #include "llvm-c/Analysis.h" // If you are getting an error here read bindings/go/README.txt
             ^~~~~~~~~~~~~~~~~~~
    1 error generated.

To continue, you'll need to build LLVM. As a first step, modify
github.com/aykevl/llvm/bindings/go/build.sh::

    cmake_flags="../../../../.. $@ -DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD=AVR -DLLVM_LINK_LLVM_DYLIB=ON"

This will enable the experimental AVR backend (for Arduino support) and will
make sure ``tinygo`` links to a shared library instead of a static library,
greatly improving link time on every rebuild. This is especially useful during
development.

The next step is actually building LLVM. This is done by running this command
inside github.com/aykevl/llvm/bindings/go::

    $ ./build.sh

This will take about an hour and require a fair bit of RAM. In fact, I would
recommend setting your ``ld`` binary to ``gold`` to speed up linking, especially
on systems with less than 16GB RAM.

After LLVM has been built, you can run an example with::

    make run-test

For a blinky example on the PCA10040 development board, do this::

    make flash-blinky2 TARGET=pca10040

Note that you will have to execute the following commands before the blinky
example will work::

    git submodule update --init
    make gen-device-nrf

You can also run a simpler blinky example (blinky1) on the Arduino::

    git submodule update --init # only required the first time
    make gen-device-avr         # only required the first time
    make flash-blinky1 TARGET=arduino
