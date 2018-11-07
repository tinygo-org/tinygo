.. _installation:

.. highlight:: none

Build tags
======================

You may have noticed that TinyGo uses a lot more build tags than is
common for Go. For example, Arduino Uno uses the following build tags:
``arduino``, ``atmega328p``, ``atmega``, ``avr5``, ``avr``, ``js``,
``wasm``. Why so many? And what does js/wasm have to do with it? It's
mostly to support the large variation in microcontrollers. Let's break
them down:

-  ``arduino``: this is the board name. Actually, it's the Arduino Uno
   so ``arduino`` is a slightly wrong name because there are other
   Arduino boards like the `Arduino
   Zero <https://store.arduino.cc/genuino-zero>`__ with a completely
   different architecture. But when talking about Arduino most people
   actually refer to the Arduino Uno so that's why it is this name.
-  ``atmega328p`` is the chip used in the Arduino Uno, the common
   `ATmega328p <https://www.microchip.com/wwwproducts/en/ATmega328P>`__.
-  ``atmega`` is the chip family. There is also the ``attiny``, for
   example, but it isn't (yet) supported by TinyGo.
-  ``avr5`` is the AVR sub-family of `many
   families <https://gcc.gnu.org/onlinedocs/gcc/AVR-Options.html>`__.
-  ``avr`` is the base instruction set architecture, which is a
   particular 8-bit Harvard-style architecture.
-  ``js``, ``wasm`` are the oddballs here. By using the standard Go
   library we have to somehow add a supported architecture to the list
   of build tags or many standard library packages will fail to compile.
   Luckily, `Go 1.11 added support for
   WebAssembly <https://github.com/golang/go/wiki/WebAssembly>`__ which
   is very useful as WebAssembly doesn't really depend on a particular
   operating system so is the closest match to a microcontroller. In
   other words, it's a hack until the Go standard library adds support
   for OS-less systems and "other" architectures (which may or may not
   happen at some point).

In general, build tags are sorted from more specific to less specific in
the .json target descriptions.
