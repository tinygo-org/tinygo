.. _usage:

.. highlight:: none

Using ``tinygo``
================


.. note::
   This page assumes you already have TinyGo installed, either :ref:`using
   Docker <docker>` or by :ref:`installing it manualling <installation>`.


Subcommands
-----------

The TinyGo tries to be similar to the main ``go`` command in usage. It consists
of the following main subcommands:

``build``
    Compile the given program. The output binary is specified using the ``-o``
    parameter. The generated file type depends on the extension:

    ``.o``
        Create a relocatable object file. You can use this option if you don't
        want to use the TinyGo build system or want to do other custom things.
    ``.ll``
        Create textual LLVM IR, after optimization. This is mainly useful for
        debugging.
    ``.bc``
        Create LLVM bitcode, after optimization. This may be useful for
        debugging or for linking into other programs using LTO.
    ``.hex``
        Create an `Intel HEX <https://en.wikipedia.org/wiki/Intel_HEX>`_ file to
        flash it to a microcontroller.
    ``.bin``
        Similar, but create a binary file.
    ``.wasm``
        Compile and link a WebAssembly file.
    (all other)
        Compile and link the program into a regular executable. For
        microcontrollers, it is common to use the .elf file extension to
        indicate a linked ELF file is generated. For Linux, it is common to
        build binaries with no extension at all.

``run``
    Run the program, either directly on the host or in an emulated environment
    (depending on ``-target``).

``flash``
    Flash the program to a microcontroller.

``gdb``
    Compile the program, optionally flash it to a microcontroller if it is a
    remote target, and drop into a GDB shell. From here you can break the
    current program (Ctrl-C), single-step, show a backtrace, etc. A debugger
    must be specified for your particular target in the target .json file and
    the required tools (like GDB for your target) must be installed as well.

``clean``
    Clean the cache directory, normally stored in ``$HOME/.cache/tinygo``. This is
    not normally needed.

``help``
    Print a short summary of the available commands, plus a list of command
    flags.


Important build options
-----------------------

There are a few flags to control how binaries are built:

``-o``
    Output filename, see the ``build`` command above.

``-target``
    Select the target you want to use. Leave it empty to compile for the host.
    Example targets:

    wasm
        WebAssembly target. Creates .wasm files that can be run in a web
        browser.
    arduino
        Compile using the experimental AVR backend to run Go programs on an
        Arduino Uno.
    microbit
        Compile programs for the `BBC micro:bit <https://microbit.org/>`_.
    qemu
        Run on a Stellaris LM3S as emulated by QEMU.

    This switch also configures the emulator, flash tool and debugger to use so
    you don't have to fiddle with those options.

    Read :ref:`supported targets <targets>` for a list of supported targets.

``-port``
    Specify the serial port used for flashing. This is used for the Arduino Uno,
    which is flashed over a serial port. It defaults to ``/dev/ttyACM0`` as that
    is the default port on Linux.

``-opt``
    Which optimization level to use. Optimization levels roughly follow standard
    ``-O`` level options like ``-O2``, ``-Os``, etc. Available optimization
    levels:

    ``-opt=0``
        Disable as much optimizations as possible. There are still a few
        optimization passes that run to make sure the program executes
        correctly, but all LLVM passes that can be disabled are disabled.
    ``-opt=1``
        Enable only a few optimization passes. In particular, this keeps the
        inliner disabled. It can be useful when you want to look at the
        generated IR but want to avoid the noise that is common in non-optimized
        code.
    ``-opt=2``
        A good optimization level for use cases not strongly limited by code
        size. Provided here for completeness. It enables most optimizations and
        will likely result in the fastest code.
    ``-opt=s``
        Like ``-opt=2``, but while being more careful about code size. It
        provides a balance between performance and code size.
    ``-opt=z`` (default)
        Like ``-opt=s``, but more aggressive about code size. This pass also
        reduces the inliner threshold by a large margin. Use this pass if you
        care a lot about code size.

``-ocd-output``
    Print output of the on-chip debugger tool (like OpenOCD) while in a ``tinygo
    gdb`` session. This can be useful to diagnose connection problems.


Miscellaneous options
---------------------

``-no-debug``
    Disable outputting debug symbols. This can be useful for WebAssembly, as
    there is no debugger for .wasm files yet and .wasm files are generally
    served directly. Avoiding debug symbols can have a big impact on generated
    binary size, reducing them by more than half.

    This is not necessary on microcontrollers because debugging symbols are not
    flashed to the microcontroller. Additionally, you will need it when you use
    ``tinygo gdb``. In general, it is recommended to include debug symbols
    unless you have a good reason not to.

    Note: while there is some support for debug symbols, only line numbers have
    been implemented so far. That means single-stepping and stacktraces work
    just fine, but no variables can be inspected.

``-size``
    Print size (``none``, ``short``, or ``full``) of the output (linked) binary.
    Note that the calculated size includes RAM reserved for the stack.

    ``none`` (default)
        Print nothing.

    ``short``
        Print size statistics, roughly like what the ``size`` binutils program
        would print but with useful flash and RAM columns::

            code    data     bss |   flash     ram
            5780     144    2132 |    5924    2276

    ``full``
        Try to determine per package how much space is used. Note that these
        calculations are merely guesses and can somethimes be way off due to
        various reasons like inlining. ::

            code  rodata    data     bss |   flash     ram | package
             876       0       4       0 |     880       4 | (bootstrap)
              38       0       0       0 |      38       0 | device/arm
               0       0       0      66 |       0      66 | machine
            2994     440     124       0 |    3558     124 | main
             948     127       4       1 |    1079       5 | runtime
            4856     567     132      67 |    5555     199 | (sum)
            5780       -     144    2132 |    5924    2276 | (all)


Compiler debugging
------------------

These options are designed to make the task of writing the compiler
significantly easier. They are seldomly useful outside of development work.

``-printir``
    Dump generated IR to the console before it is optimized.

``-dumpssa``
    Dump Go SSA to the console while the program is being compiled. This also
    includes the SSA of package initializers while they are being interpreted.
