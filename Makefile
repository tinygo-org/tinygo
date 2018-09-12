
# aliases
all: tgo
tgo: build/tgo

.PHONY: all tgo run-test run-blinky run-blinky2 clean fmt gen-device gen-device-nrf gen-device-avr

TARGET ?= unix

ifeq ($(TARGET),unix)
# Regular *nix system.
SIZE = size

else ifeq ($(TARGET),pca10040)
# PCA10040: nRF52832 development board
SIZE = arm-none-eabi-size
OBJCOPY = arm-none-eabi-objcopy
TGOFLAGS += -target $(TARGET)

else ifeq ($(TARGET),arduino)
SIZE = avr-size
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
else ifeq ($(TARGET),arduino)
flash-%: build/%.hex
	avrdude -c arduino -p atmega328p -P /dev/ttyACM0 -U flash:w:$<
endif

clean:
	@rm -rf build

fmt:
	@go fmt . ./src/examples/test
	@go fmt ./src/runtime/*.go

gen-device: gen-device-nrf gen-device-avr

gen-device-nrf:
	./tools/gen-device-svd.py lib/nrfx/mdk/ src/device/nrf/
	go fmt ./src/device/nrf

gen-device-avr:
	./tools/gen-device-avr.py lib/avr/packs/atmega src/device/avr/
	go fmt ./src/device/avr


# Build the Go compiler.
build/tgo: *.go
	@mkdir -p build
	go build -o build/tgo -i .

# Binary that can run on the host.
build/%: src/examples/% src/examples/%/*.go build/tgo src/runtime/*.go
	./build/tgo build $(TGOFLAGS) -o $@ $(subst src/,,$<)
	@$(SIZE) $@

# ELF file that can run on a microcontroller.
build/%.elf: src/examples/% src/examples/%/*.go build/tgo src/runtime/*.go
	./build/tgo build $(TGOFLAGS) -o $@ $(subst src/,,$<)
	@$(SIZE) $@

# Convert executable to Intel hex file (for flashing).
build/%.hex: build/%.elf
	$(OBJCOPY) -O ihex $^ $@
