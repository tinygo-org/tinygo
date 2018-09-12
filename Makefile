
# aliases
all: tgo
tgo: build/tgo

.PHONY: all tgo run-test run-blinky run-blinky2 clean fmt gen-device gen-device-nrf

CFLAGS = -Wall -Werror -Os -fno-exceptions -ffunction-sections -fdata-sections $(LLFLAGS)
CFLAGS += -fno-exceptions -fno-unwind-tables # Avoid .ARM.exidx etc.

TARGET ?= unix

ifeq ($(TARGET),unix)
# Regular *nix system.
LD = clang
SIZE = size

else ifeq ($(TARGET),pca10040)
# PCA10040: nRF52832 development board
LD = arm-none-eabi-ld -T arm.ld --gc-sections
SIZE = arm-none-eabi-size
OBJCOPY = arm-none-eabi-objcopy
LLFLAGS += -target armv7m-none-eabi
TGOFLAGS += -target $(TARGET)
CFLAGS += -I$(CURDIR)/lib/CMSIS/CMSIS/Include
CFLAGS += -DNRF52832_XXAA
CFLAGS += -Wno-uninitialized
OBJ += build/nrfx_system_nrf52.o
OBJ += build/nrfx_startup_nrf51.o # TODO nrf52, see https://bugs.llvm.org/show_bug.cgi?id=31601

else ifeq ($(TARGET),arduino)
AS = avr-as -mmcu=atmega328p
LD = avr-ld -T avr.ld --gc-sections
SIZE = avr-size
OBJCOPY = avr-objcopy
LLFLAGS += -target avr8--
TGOFLAGS += -target $(TARGET)
OBJ += build/avr.o

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

gen-device: gen-device-nrf

gen-device-nrf:
	./gen-device-svd.py lib/nrfx/mdk/ src/device/nrf/
	go fmt ./src/device/nrf

gen-device-avr:
	./gen-device-avr.py lib/avr/packs/atmega src/device/avr/
	go fmt ./src/device/avr


# Build the Go compiler.
build/tgo: *.go
	@mkdir -p build
	go build -o build/tgo -i .

# Build IR with the Go compiler.
build/%.o: src/examples/% src/examples/%/*.go build/tgo src/runtime/*.go
	./build/tgo build $(TGOFLAGS) -o $@ $(subst src/,,$<)

# Compile system_* file for the nRF.
build/nrfx_%.o: lib/nrfx/mdk/%.c
	@mkdir -p build
	clang $(CFLAGS) -c -o $@ $^

# Compile startup_* file for the nRF.
build/nrfx_%.o: lib/nrfx/mdk/gcc_%.S
	@mkdir -p build
	clang $(CFLAGS) -D__STARTUP_CLEAR_BSS -c -o $@ $^

build/%.o: %.S
	$(AS) -o $@ $^

# Generate output ELF executable.
build/%: build/%.o $(OBJ)
	$(LD) -o $@ $^
	$(SIZE) $@

# Generate output ELF for use in objcopy (on a microcontroller).
build/%.elf: build/%.o $(OBJ)
	$(LD) -o $@ $^
	$(SIZE) $@

# Convert executable to Intel hex file (for flashing).
build/%.hex: build/%.elf
	$(OBJCOPY) -O ihex $^ $@
