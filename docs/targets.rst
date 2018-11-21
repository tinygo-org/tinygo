.. _targets:

.. |br| raw:: html

    <br>

Supported targets
=================

TinyGo makes it easy to add new targets. If your target isn't listed here,
please raise an issue in the `issue tracker
<https://github.com/aykevl/tinygo/issues>`_.


POSIX-like
----------

Only Linux is supported at the moment, but it should be trivial to add support
for more POSIX-like systems.


ARM / Cortex-M
--------------

Cortex-M processors are well supported. There is support for multiple chips and
the backend appears to be stable. In fact, it uses the same underlying
technology (LLVM) as the proprietary ARM compiler for code generation.

  * `BBC micro:bit <https://microbit.org/>`_ (`nRF51822
    <https://www.nordicsemi.com/eng/Products/Bluetooth-low-energy/nRF51822>`_)
  * `Nordic PCA10031
    <https://www.nordicsemi.com/eng/Products/nRF51-Dongle>`_
    (`nRF51422
    <https://www.nordicsemi.com/eng/Products/ANT/nRF51422>`_)
  * `Nordic PCA10040
    <https://www.nordicsemi.com/eng/Products/Bluetooth-low-energy/nRF52-DK>`_
    (`nRF52832
    <https://www.nordicsemi.com/eng/Products/Bluetooth-low-energy/nRF52832>`_)
  * `nRF52840-MDK <https://wiki.makerdiary.com/nrf52840-mdk/>`_ (`nRF52840
    <https://www.nordicsemi.com/eng/Products/nRF52840>`_)
  * `QEMU <https://wiki.qemu.org/Documentation/Platforms/ARM>`_ (`LM3S6965
    <http://www.ti.com/product/LM3S6965>`_) |br|
    This target is supported only for testing purposes. It has not been tested
    on real hardware.


AVR
---

Note: the AVR backend of LLVM is still experimental so you may encounter bugs.

  * `Arduino Uno <https://store.arduino.cc/arduino-uno-rev3>`_ (`ATmega328p
    <https://www.microchip.com/wwwproducts/en/ATmega328p>`_)
  * `Digispark <http://digistump.com/products/1>`_ (`ATtiny85
    <https://www.microchip.com/wwwproducts/en/ATtiny85>`_) |br|
    Very limited support at the moment.


WebAssembly
-----------

WebAssembly support is relatively new but appears to be stable.


.. note::
   Support for the ESP8266/ESP32 chips will take a lot of work if they ever get
   support. See :ref:`this FAQ entry <faq-esp>` for details.
