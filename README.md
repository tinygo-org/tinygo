# TinyGo - Go compiler for small places

[![Linux](https://github.com/tinygo-org/tinygo/actions/workflows/linux.yml/badge.svg?branch=dev)](https://github.com/tinygo-org/tinygo/actions/workflows/linux.yml) [![macOS](https://github.com/tinygo-org/tinygo/actions/workflows/build-macos.yml/badge.svg?branch=dev)](https://github.com/tinygo-org/tinygo/actions/workflows/build-macos.yml) [![Windows](https://github.com/tinygo-org/tinygo/actions/workflows/windows.yml/badge.svg?branch=dev)](https://github.com/tinygo-org/tinygo/actions/workflows/windows.yml) [![Docker](https://github.com/tinygo-org/tinygo/actions/workflows/docker.yml/badge.svg?branch=dev)](https://github.com/tinygo-org/tinygo/actions/workflows/docker.yml) [![CircleCI](https://circleci.com/gh/tinygo-org/tinygo/tree/dev.svg?style=svg)](https://circleci.com/gh/tinygo-org/tinygo/tree/dev)

TinyGo is a Go compiler intended for use in small places such as microcontrollers, WebAssembly (Wasm), and command-line tools.

It reuses libraries used by the [Go language tools](https://golang.org/pkg/go/) alongside [LLVM](http://llvm.org) to provide an alternative way to compile programs written in the Go programming language.

Here is an example program that blinks the built-in LED when run directly on any supported board with onboard LED:

```go
package main

import (
    "machine"
    "time"
)

func main() {
    led := machine.LED
    led.Configure(machine.PinConfig{Mode: machine.PinOutput})
    for {
        led.Low()
        time.Sleep(time.Millisecond * 1000)

        led.High()
        time.Sleep(time.Millisecond * 1000)
    }
}
```

The above program can be compiled and run without modification on an Arduino Uno, an Adafruit ItsyBitsy M0, or any of the supported boards that have a built-in LED, just by setting the correct TinyGo compiler target. For example, this compiles and flashes an Arduino Uno:

```shell
tinygo flash -target arduino examples/blinky1
```

## Installation

See the [getting started instructions](https://tinygo.org/getting-started/) for information on how to install TinyGo, as well as how to run the TinyGo compiler using our Docker container.

## Supported boards/targets

You can compile TinyGo programs for microcontrollers, WebAssembly and Linux.

The following 82 microcontroller boards are currently supported:

* [Adafruit Circuit Playground Bluefruit](https://www.adafruit.com/product/4333)
* [Adafruit Circuit Playground Express](https://www.adafruit.com/product/3333)
* [Adafruit CLUE](https://www.adafruit.com/product/4500)
* [Adafruit Feather M0](https://www.adafruit.com/product/2772)
* [Adafruit Feather M4](https://www.adafruit.com/product/3857)
* [Adafruit Feather M4 CAN](https://www.adafruit.com/product/4759)
* [Adafruit Feather nRF52840 Express](https://www.adafruit.com/product/4062)
* [Adafruit Feather nRF52840 Sense](https://www.adafruit.com/product/4516)
* [Adafruit Feather RP2040](https://www.adafruit.com/product/4884)
* [Adafruit Feather STM32F405 Express](https://www.adafruit.com/product/4382)
* [Adafruit Grand Central M4](https://www.adafruit.com/product/4064)
* [Adafruit ItsyBitsy M0](https://www.adafruit.com/product/3727)
* [Adafruit ItsyBitsy M4](https://www.adafruit.com/product/3800)
* [Adafruit ItsyBitsy nRF52840](https://www.adafruit.com/product/4481)
* [Adafruit MacroPad RP2040](https://www.adafruit.com/product/5100)
* [Adafruit Matrix Portal M4](https://www.adafruit.com/product/4745)
* [Adafruit Metro M4 Express Airlift](https://www.adafruit.com/product/4000)
* [Adafruit PyBadge](https://www.adafruit.com/product/4200)
* [Adafruit PyGamer](https://www.adafruit.com/product/4242)
* [Adafruit PyPortal](https://www.adafruit.com/product/4116)
* [Adafruit QT Py](https://www.adafruit.com/product/4600)
* [Adafruit Trinket M0](https://www.adafruit.com/product/3500)
* [Arduino Mega 1280](https://www.arduino.cc/en/Main/arduinoBoardMega/)
* [Arduino Mega 2560](https://store.arduino.cc/arduino-mega-2560-rev3)
* [Arduino MKR1000](https://store.arduino.cc/arduino-mkr1000-wifi)
* [Arduino MKR WiFi 1010](https://store.arduino.cc/usa/mkr-wifi-1010)
* [Arduino Nano](https://store.arduino.cc/arduino-nano)
* [Arduino Nano 33 BLE](https://store.arduino.cc/nano-33-ble)
* [Arduino Nano 33 BLE Sense](https://store.arduino.cc/nano-33-ble-sense)
* [Arduino Nano 33 IoT](https://store.arduino.cc/nano-33-iot)
* [Arduino Nano RP2040 Connect](https://store.arduino.cc/nano-rp2040-connect)
* [Arduino Uno](https://store.arduino.cc/arduino-uno-rev3)
* [Arduino Zero](https://store.arduino.cc/usa/arduino-zero)
* [BBC micro:bit](https://microbit.org/)
* [BBC micro:bit v2](https://microbit.org/new-microbit/)
* [blues wireless Swan](https://blues.io/products/swan/)
* [Digispark](http://digistump.com/products/1)
* [Dragino LoRaWAN GPS Tracker LGT-92](http://www.dragino.com/products/lora-lorawan-end-node/item/142-lgt-92.html)
* [ESP32 - Core board](https://www.espressif.com/en/products/socs/esp32)
* [ESP32 - mini32](https://www.espressif.com/en/products/socs/esp32)
* [ESP8266 - d1mini](https://www.espressif.com/en/products/socs/esp8266)
* [ESP8266 - NodeMCU](https://www.espressif.com/en/products/socs/esp8266)
* [Game Boy Advance](https://en.wikipedia.org/wiki/Game_Boy_Advance)
* [M5Stack](https://docs.m5stack.com/en/core/basic)
* [M5Stack Core2](https://shop.m5stack.com/products/m5stack-core2-esp32-iot-development-kit)
* [M5Stamp C3](https://docs.m5stack.com/en/core/stamp_c3)
* [Makerdiary nRF52840-MDK](https://wiki.makerdiary.com/nrf52840-mdk/)
* [Makerdiary nRF52840-MDK USB Dongle](https://wiki.makerdiary.com/nrf52840-mdk-usb-dongle/)
* [Microchip SAM E54 Xplained Pro](https://www.microchip.com/developmenttools/productdetails/atsame54-xpro)
* [nice!nano](https://docs.nicekeyboards.com/#/nice!nano/)
* [Nintendo Switch](https://www.nintendo.com/switch/)
* [Nordic Semiconductor PCA10031](https://www.nordicsemi.com/eng/Products/nRF51-Dongle)
* [Nordic Semiconductor PCA10040](https://www.nordicsemi.com/eng/Products/Bluetooth-low-energy/nRF52-DK)
* [Nordic Semiconductor PCA10056](https://www.nordicsemi.com/Software-and-Tools/Development-Kits/nRF52840-DK)
* [Nordic Semiconductor pca10059](https://www.nordicsemi.com/Software-and-tools/Development-Kits/nRF52840-Dongle)
* [Particle Argon](https://docs.particle.io/datasheets/wi-fi/argon-datasheet/)
* [Particle Boron](https://docs.particle.io/datasheets/cellular/boron-datasheet/)
* [Particle Xenon](https://docs.particle.io/datasheets/discontinued/xenon-datasheet/)
* [Phytec reel board](https://www.phytec.eu/product-eu/internet-of-things/reelboard/)
* [PineTime DevKit](https://www.pine64.org/pinetime/)
* [PJRC Teensy 3.6](https://www.pjrc.com/store/teensy36.html)
* [PJRC Teensy 4.0](https://www.pjrc.com/store/teensy40.html)
* [PJRC Teensy 4.1](https://www.pjrc.com/store/teensy41.html)
* [ProductivityOpen P1AM-100](https://facts-engineering.github.io/modules/P1AM-100/P1AM-100.html)
* [Raspberry Pi Pico](https://www.raspberrypi.org/products/raspberry-pi-pico/)
* [Raytac MDBT50Q-RX Dongle (with TinyUF2 bootloader)](https://www.adafruit.com/product/5199)
* [Seeed Seeeduino XIAO](https://www.seeedstudio.com/Seeeduino-XIAO-Arduino-Microcontroller-SAMD21-Cortex-M0+-p-4426.html)
* [Seeed LoRa-E5 Development Kit](https://www.seeedstudio.com/LoRa-E5-Dev-Kit-p-4868.html)
* [Seeed Sipeed MAix BiT](https://www.seeedstudio.com/Sipeed-MAix-BiT-for-RISC-V-AI-IoT-p-2872.html)
* [Seeed Wio Terminal](https://www.seeedstudio.com/Wio-Terminal-p-4509.html)
* [SiFIve HiFive1 Rev B](https://www.sifive.com/boards/hifive1)
* [ST Micro "Nucleo" F103RB](https://www.st.com/en/evaluation-tools/nucleo-f103rb.html)
* [ST Micro "Nucleo" F722ZE](https://www.st.com/en/evaluation-tools/nucleo-f722ze.html)
* [ST Micro "Nucleo" L031K6](https://www.st.com/ja/evaluation-tools/nucleo-l031k6.html)
* [ST Micro "Nucleo" L432KC](https://www.st.com/ja/evaluation-tools/nucleo-l432kc.html)
* [ST Micro "Nucleo" L552ZE](https://www.st.com/en/evaluation-tools/nucleo-l552ze-q.html)
* [ST Micro "Nucleo" WL55JC](https://www.st.com/en/evaluation-tools/nucleo-wl55jc.html)
* [ST Micro STM32F103XX "Bluepill"](https://stm32-base.org/boards/STM32F103C8T6-Blue-Pill)
* [ST Micro STM32F407 "Discovery"](https://www.st.com/en/evaluation-tools/stm32f4discovery.html)
* [ST Micro STM32F469 "Discovery"](https://www.st.com/content/st_com/en/products/evaluation-tools/product-evaluation-tools/mcu-mpu-eval-tools/stm32-mcu-mpu-eval-tools/stm32-discovery-kits/32f469idiscovery.html)
* [X9 Pro smartwatch](https://github.com/curtpw/nRF5x-device-reverse-engineering/tree/master/X9-nrf52832-activity-tracker/)
* [The Things Industries Generic Node Sensor Edition](https://www.genericnode.com/docs/sensor-edition/)

For more information, see [this list of boards](https://tinygo.org/microcontrollers/). Pull requests for additional support are welcome!

## Currently supported features:

For a description of currently supported Go language features, please see [https://tinygo.org/lang-support/](https://tinygo.org/lang-support/).

## Documentation

Documentation is located on our web site at [https://tinygo.org/](https://tinygo.org/).

You can find the web site code at [https://github.com/tinygo-org/tinygo-site](https://github.com/tinygo-org/tinygo-site).

## Getting help

If you're looking for a more interactive way to discuss TinyGo usage or
development, we have a [#TinyGo channel](https://gophers.slack.com/messages/CDJD3SUP6/)
on the [Gophers Slack](https://gophers.slack.com).

If you need an invitation for the Gophers Slack, you can generate one here which
should arrive fairly quickly (under 1 min): https://invite.slack.golangbridge.org

## Contributing

Your contributions are welcome!

Please take a look at our [Contributing](https://tinygo.org/docs/guides/contributing/) page on our web site for details.

## Project Scope

Goals:

* Have very small binary sizes. Don't pay for what you don't use.
* Support for most common microcontroller boards.
* Be usable on the web using WebAssembly.
* Good CGo support, with no more overhead than a regular function call.
* Support most standard library packages and compile most Go code without modification.

Non-goals:

* Be efficient while using zillions of goroutines. However, good goroutine support is certainly a goal.
* Be as fast as `gc`. However, LLVM will probably be better at optimizing certain things so TinyGo might actually turn out to be faster for number crunching.
* Be able to compile every Go program out there.

## Why this project exists

> We never expected Go to be an embedded language and so its got serious problems...

-- Rob Pike, [GopherCon 2014 Opening Keynote](https://www.youtube.com/watch?v=VoS7DsT1rdM&feature=youtu.be&t=2799)

TinyGo is a project to bring Go to microcontrollers and small systems with a single processor core. It is similar to [emgo](https://github.com/ziutek/emgo) but a major difference is that we want to keep the Go memory model (which implies garbage collection of some sort). Another difference is that TinyGo uses LLVM internally instead of emitting C, which hopefully leads to smaller and more efficient code and certainly leads to more flexibility.

The original reasoning was: if [Python](https://micropython.org/) can run on microcontrollers, then certainly [Go](https://golang.org/) should be able to run on even lower level micros.

## License

This project is licensed under the BSD 3-clause license, just like the [Go project](https://golang.org/LICENSE) itself.

Some code has been copied from the LLVM project and is therefore licensed under [a variant of the Apache 2.0 license](http://releases.llvm.org/11.0.0/LICENSE.TXT). This has been clearly indicated in the header of these files.

Some code has been copied and/or ported from Paul Stoffregen's Teensy libraries and is therefore licensed under PJRC's license. This has been clearly indicated in the header of these files.
