.. targets:

.. |br| raw:: html

    <br>

Supported targets
=================

The following boards are currently supported:

  * `Arduino Uno <https://store.arduino.cc/arduino-uno-rev3>`_ (`ATmega328p
    <https://www.microchip.com/wwwproducts/en/ATmega328p>`_) |br|
    Note: the AVR backend of LLVM is still experimental so you may encounter
    bugs.
  * `BBC micro:bit <https://microbit.org/>`_ (`nRF51822
    <https://www.nordicsemi.com/eng/Products/Bluetooth-low-energy/nRF51822>`_)
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


.. note::
   Support for the ESP8266/ESP32 chips will take a lot of work if they ever get
   support. See :ref:`this FAQ entry <faq-esp>` for details.
