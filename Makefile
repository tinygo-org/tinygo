
# aliases
all: tinygo
tinygo: build/tinygo

.PHONY: all tinygo run-test run-blinky run-blinky2 clean fmt gen-device gen-device-nrf gen-device-avr

TARGET ?= unix

ifeq ($(TARGET),unix)
# Regular *nix system.

else ifeq ($(TARGET),pca10040)
# PCA10040: nRF52832 development board
OBJCOPY = arm-none-eabi-objcopy
TGOFLAGS += -target $(TARGET)

else ifeq ($(TARGET),microbit)
# BBC micro:bit
OBJCOPY = arm-none-eabi-objcopy
TGOFLAGS += -target $(TARGET)

else ifeq ($(TARGET),reelboard)
# reel board
OBJCOPY = arm-none-eabi-objcopy
TGOFLAGS += -target $(TARGET)

else ifeq ($(TARGET),bluepill)
# "blue pill" development board
# See: https://wiki.stm32duino.com/index.php?title=Blue_Pill
OBJCOPY = arm-none-eabi-objcopy
TGOFLAGS += -target $(TARGET)

else ifeq ($(TARGET),arduino)
OBJCOPY = avr-objcopy
TGOFLAGS += -target $(TARGET)

else
$(error Unknown target)

endif



run-test: build/test
	./build/test

run-blinky: run-blinky2
run-blinky2: build/blinky2
	./build/blinky2

ifeq ($(TARGET),pca10040)
flash-%: build/%.hex
	nrfjprog -f nrf52 --sectorerase --program $< --reset
else ifeq ($(TARGET),microbit)
flash-%: build/%.hex
	openocd -f interface/cmsis-dap.cfg -f target/nrf51.cfg -c 'program $< reset exit'
else ifeq ($(TARGET),reelboard)
flash-%: build/%.hex
	openocd -f interface/cmsis-dap.cfg -f target/nrf51.cfg -c 'program $< reset exit'
else ifeq ($(TARGET),arduino)
flash-%: build/%.hex
	avrdude -c arduino -p atmega328p -P /dev/ttyACM0 -U flash:w:$<
else ifeq ($(TARGET),bluepill)
flash-%: build/%.hex
	openocd -f interface/stlink-v2.cfg -f target/stm32f1x.cfg -c 'program $< reset exit'
endif

clean:
	@rm -rf build

fmt:
	@go fmt . ./compiler ./interp ./loader ./ir ./src/device/arm ./src/examples/* ./src/machine ./src/runtime ./src/sync
	@go fmt ./testdata/*.go

test:
	@go test -v .

gen-device: gen-device-avr gen-device-nrf gen-device-sam gen-device-stm32

gen-device-avr:
	./tools/gen-device-avr.py lib/avr/packs/atmega src/device/avr/
	./tools/gen-device-avr.py lib/avr/packs/tiny src/device/avr/
	go fmt ./src/device/avr

gen-device-nrf:
	./tools/gen-device-svd.py lib/nrfx/mdk/ src/device/nrf/ --source=https://github.com/NordicSemiconductor/nrfx/tree/master/mdk
	go fmt ./src/device/nrf

gen-device-sam:
	./tools/gen-device-svd.py lib/cmsis-svd/data/Atmel/ src/device/sam/ --source=https://github.com/posborne/cmsis-svd/tree/master/data/Atmel
	go fmt ./src/device/sam

gen-device-stm32:
	./tools/gen-device-svd.py lib/cmsis-svd/data/STMicro/ src/device/stm32/ --source=https://github.com/posborne/cmsis-svd/tree/master/data/STMicro
	go fmt ./src/device/stm32

# Build the Go compiler.
tinygo:
	@mkdir -p build
	go build -o build/tinygo .

# Binary that can run on the host.
build/%: src/examples/% src/examples/%/*.go build/tinygo src/runtime/*.go
	./build/tinygo build $(TGOFLAGS) -size=short -o $@ $(subst src/,,$<)

# ELF file that can run on a microcontroller.
build/%.elf: src/examples/% src/examples/%/*.go build/tinygo src/runtime/*.go
	./build/tinygo build $(TGOFLAGS) -size=short -o $@ $(subst src/,,$<)

# Convert executable to Intel hex file (for flashing).
build/%.hex: build/%.elf
	$(OBJCOPY) -O ihex $^ $@
