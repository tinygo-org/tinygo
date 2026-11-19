# Generate microcontroller-specific sources from SVD files.

GEN_DEVICE_TARGETS = \
	gen-device-atmega \
	gen-device-attiny \
	gen-device-esp \
	gen-device-nrf \
	gen-device-sam \
	gen-device-sifive \
	gen-device-kendryte \
	gen-device-nxp \
	gen-device-stm32 \
	gen-device-rp \
	gen-device-renesas \
	gen-device-py32

.PHONY: gen-device $(GEN_DEVICE_TARGETS) gen-target-py32

gen-device: $(GEN_DEVICE_TARGETS)

build/gen-device-%: ./tools/gen-device-%/*.go
	$(GO) build -o $@ ./tools/gen-device-$*/

# ATmega

DEVICES_AVR_MEGA = 1280 1284P 2560 328PB 328P 32U4

gen-device-atmega: $(foreach m,$(DEVICES_AVR_MEGA),src/device/avr/ATmega$(m).go src/device/avr/ATmega$(m).s src/device/avr/ATmega$(m).ld)

src/device/avr/ATmega%.go src/device/avr/ATmega%.s src/device/avr/ATmega%.ld: lib/avr/packs/atmega/ATmega%.atdf build/gen-device-avr
	build/gen-device-avr $< src/device/avr/ATmega$*.go src/device/avr/ATmega$*.s src/device/avr/ATmega$*.ld

# ATtiny

DEVICES_AVR_TINY = 1616 85

gen-device-attiny: $(foreach m,$(DEVICES_AVR_TINY),src/device/avr/ATtiny$(m).go src/device/avr/ATtiny$(m).s src/device/avr/ATtiny$(m).ld)

src/device/avr/ATtiny%.go src/device/avr/ATtiny%.s src/device/avr/ATtiny%.ld: lib/avr/packs/tiny/ATtiny%.atdf build/gen-device-avr
	build/gen-device-avr $< src/device/avr/ATtiny$*.go src/device/avr/ATtiny$*.s src/device/avr/ATtiny$*.ld

# Espressif

# Both Espressif and Espressif-Community define esp32.
# The Espressif directory takes priority. 

DEVICES_ESP = 32c3 32c6 32h2 32s3 32 8266

gen-device-esp: $(foreach m,$(DEVICES_ESP),src/device/esp/esp$(m).go)

src/device/esp/%.go: lib/cmsis-svd/data/Espressif/%.svd build/gen-device-svd
	build/gen-device-svd -url=https://github.com/posborne/cmsis-svd/tree/master/data/Espressif -interrupts=software $< $@

src/device/esp/%.go: lib/cmsis-svd/data/Espressif-Community/%.svd build/gen-device-svd
	build/gen-device-svd -url=https://github.com/posborne/cmsis-svd/tree/master/data/Espressif-Community -interrupts=software $< $@

# NRF

DEVICES_NRF = 51 52833 52840 52

gen-device-nrf: $(foreach m,$(DEVICES_NRF),src/device/nrf/nrf$(m).go src/device/nrf/nrf$(m).s)

src/device/nrf/nrf%.go src/device/nrf/nrf%.s: lib/nrfx/mdk/nrf%.svd build/gen-device-svd
	build/gen-device-svd -url=https://github.com/NordicSemiconductor/nrfx/tree/master/mdk -asm=src/device/nrf/nrf$*.s $< src/device/nrf/nrf$*.go

# NXP

DEVICES_NXP = MIMXRT1062 MK66F18

gen-device-nxp: $(foreach m,$(DEVICES_NXP),src/device/nxp/$(m).go src/device/nxp/$(m).s)

src/device/nxp/%.go src/device/nxp/%.s: lib/cmsis-svd/data/NXP/%.svd build/gen-device-svd
	build/gen-device-svd -url=https://github.com/posborne/cmsis-svd/tree/master/data/NXP -asm=src/device/nxp/$*.s $< src/device/nxp/$*.go

# SAM

DEVICES_SAM = D21E18A D21G18A D51G19A D51J19A D51J20A D51P19A D51P20A D51J19A E51J19A E54P20A

gen-device-sam: $(foreach m,$(DEVICES_SAM),src/device/sam/ATSAM$(m).go src/device/sam/ATSAM$(m).s)

src/device/sam/ATSAM%.go src/device/sam/ATSAM%.s: lib/cmsis-svd/data/Atmel/ATSAM%.svd build/gen-device-svd
	build/gen-device-svd -url=https://github.com/posborne/cmsis-svd/tree/master/data/Atmel -asm=src/device/sam/ATSAM$*.s $< src/device/sam/ATSAM$*.go

# SiFive

DEVICES_SIFIVE = e310x

gen-device-sifive: $(foreach m,$(DEVICES_SIFIVE),src/device/sifive/$(m).go)

src/device/sifive/%.go: lib/cmsis-svd/data/SiFive-Community/%.svd build/gen-device-svd
	build/gen-device-svd -url=https://github.com/posborne/cmsis-svd/tree/master/data/SiFive-Community -interrupts=software $< $@

# Kendryte

DEVICES_KENDRYTE = 210

gen-device-kendryte: $(foreach m,$(DEVICES_KENDRYTE),src/device/kendryte/k$(m).go)

src/device/kendryte/k%.go: lib/cmsis-svd/data/Kendryte-Community/k%.svd build/gen-device-svd
	build/gen-device-svd -url=https://github.com/posborne/cmsis-svd/tree/master/data/Kendryte-Community -interrupts=software $< $@

# STM32

DEVICES_STM32 = f103 f401 f405 f407 f469 f722 g0b1 h753 l4r5 l0x1 l0x2 l4x2 l4x5 l4x6 l552 u031 u585 wl5x_cm4 wle5

gen-device-stm32: $(foreach m,$(DEVICES_STM32),src/device/stm32/stm32$(m).go src/device/stm32/stm32$(m).s)

src/device/stm32/stm32%.go src/device/stm32/stm32%.s: lib/stm32-svd/svd/stm32%.svd build/gen-device-svd
	build/gen-device-svd -url=https://github.com/tinygo-org/stm32-svd -asm=src/device/stm32/stm32$*.s $< src/device/stm32/stm32$*.go

# Raspberry Pi

DEVICES_RP = 2040 2350

gen-device-rp: $(foreach m,$(DEVICES_RP),src/device/rp/rp$(m).go src/device/rp/rp$(m).s)

src/device/rp/rp%.go src/device/rp/rp%.s: lib/cmsis-svd/data/RaspberryPi/rp%.svd build/gen-device-svd
	build/gen-device-svd -url=https://github.com/posborne/cmsis-svd/tree/master/data/RaspberryPi -asm=src/device/rp/rp$*.s $< src/device/rp/rp$*.go

# Renesas

DEVICES_RENESAS = R7FA4M1AB

gen-device-renesas: $(foreach m,$(DEVICES_RENESAS),src/device/renesas/$(m).go src/device/renesas/$(m).s)

src/device/renesas/%.go src/device/renesas/%.s: lib/cmsis-svd/data/Renesas/%.svd build/gen-device-svd
	build/gen-device-svd -url=https://github.com/cmsis-svd/cmsis-svd-data/tree/master/data/Renesas -asm=src/device/renesas/$*.s $< src/device/renesas/$*.go

# Puya PY32

# Do we really want to support *all* of these chips?
DEVICES_PY32 = \
	e407xx \
	f001cxx f001xx \
	f002axx f002bxx f002cxx f002xxx f002zxx \
	f003xx \
	f021xx \
	f030xx f031xx f032xx \
	f040cxx f040epxx f040exx f040xx \
	f071cxx f071exx f071xx f072cxx f072exx f072xx \
	f403xx f410xx \
	l020xx l090xx \
	m010xx m020xx \
	m030xx m031xx \
	m070cxx m070xx \
	md310xx md320xx \
	md410xx md420xx md430xx \
	t020xx t090xx t092xx

gen-device-py32: $(foreach m,$(DEVICES_PY32),src/device/py32/py32$(m).go src/device/py32/py32$(m).s)

src/device/py32/py32%.go src/device/py32/py32%.s: lib/py32-svd/svd/py32%.svd build/gen-device-svd
	build/gen-device-svd -url=https://github.com/tinygo-org/py32-svd -asm=src/device/py32/py32$*.s $< src/device/py32/py32$*.go

gen-target-py32: ## Generate PY32 target and linker definitions
	$(GO) run ./tools/gen-py32-targets
