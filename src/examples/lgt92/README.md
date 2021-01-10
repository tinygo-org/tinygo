
Some examples for testing Dragino LGT92 devices (UARTs, SPI, LED, GPS, LORA ...) 


# Flashing LGT92 

  # With STLINKv2

You can flash LGT92 with the provided USB cable adapter and STLinkV2 adapter:

|LGT92 Cable|ST Link v2|
|-|-|
|RED| +5V|
|WHITE|SWDIO|
|GREEN|SWCLK|
|BLACK|GND|

  # With Black Magic

... Or with a BlackMagic adapter: 

|LGT92 Cable|BlackMagic|
|-|-|
|RED|+5V|
|WHITE|PB14(SWDIO)|
|GREEN|PA5(SWCLK) |
|BLACK|GND|


# LGT92 console uart

After flashing, the firmware turns SWD LGT92 pins (green/white wires) into a UART.

You can connect a UART interface to these pins to read the messages. 

(Note : You can't connect UART interface and SWD Flasher at the same time)


|LGT92 Cable|USART adapter|
|-|-|
|WHITE| TX  |
|GREEN| RX |

If you use black magic extra uart: 

|LGT92 Cable|Black Magic|
|-|-|
|WHITE| PA2 (TX)  |
|GREEN| PA3 (RX) |


# Building examples


|File|Description|
|-|-|
|01_gpio_int_button.go|Handle button GPIO with interrupt|
|02_simple_blink.go| Red led will blink every seconds|
|03_gps_read.go|Read NMEA sentences from UART|


Example:

```
tinygo build --size=full --target=lgt92 -o main.elf gpio_int1.go
```

# How to flash

Either use tinygo flash command: 

```
tinygo flash --target=lgt92  01_gpio_int_button.go
```

Or  do it 'by hand': 

```
openocd -f /usr/share/openocd/scripts/interface/stlink-v2.cfg -c "transport select hla_swd" -f /usr/share/openocd/scripts/target/stm32l0.cfg -c init -c 'program main.elf verify reset exit'
```

Or even with BlackMagic interface : 

```
tinygo build -size=full -target=lgt92 -opt=z -o 03_gps.elf 03_gps.go

arm-none-eabi-gdb -nx --batch -ex "target extended-remote /dev/ttyACM1" -ex "monitor swdp_scan" -ex "attach 1" -ex load -ex compare-sections -ex kill 03_gps
```
