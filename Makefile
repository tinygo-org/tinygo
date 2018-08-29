
# aliases
all: tgo
tgo: build/tgo

.PHONY: all tgo run-hello run-blinky clean fmt gen-device gen-device-nrf

# Custom LLVM toolchain.
LLVM := $(shell go env GOPATH)/src/github.com/aykevl/llvm/bindings/go/llvm/workdir/llvm_build/bin/
LINK = $(LLVM)llvm-link
LLC = $(LLVM)llc
LLAS = $(LLVM)llvm-as
OPT = $(LLVM)opt

CFLAGS = -Wall -Werror -Os -fno-exceptions -flto -ffunction-sections -fdata-sections $(LLFLAGS)
CFLAGS += -fno-exceptions -fno-unwind-tables # Avoid .ARM.exidx etc.

RUNTIME_PARTS = build/runtime.bc

TARGET ?= unix

ifeq ($(TARGET),unix)
# Regular *nix system.
GCC = gcc
LD = clang
SIZE = size

else ifeq ($(TARGET),pca10040)
# PCA10040: nRF52832 development board
GCC = arm-none-eabi-gcc
LD = arm-none-eabi-ld -T arm.ld --gc-sections
SIZE = arm-none-eabi-size
OBJCOPY = arm-none-eabi-objcopy
LLFLAGS += -target armv7m-none-eabi
TGOFLAGS += -target $(TARGET)
CFLAGS += -I$(CURDIR)/src/runtime
CFLAGS += -I$(CURDIR)/lib/nrfx
CFLAGS += -I$(CURDIR)/lib/nrfx/mdk
CFLAGS += -I$(CURDIR)/lib/CMSIS/CMSIS/Include
CFLAGS += -DNRF52832_XXAA
CFLAGS += -Wno-uninitialized
RUNTIME_PARTS += build/runtime_nrf.bc
RUNTIME_PARTS += build/nrfx_system_nrf52.bc
OBJ += build/nrfx_startup_nrf51.o # TODO nrf52, see https://bugs.llvm.org/show_bug.cgi?id=31601

else ifeq ($(TARGET),arduino)
GCC = avr-gcc
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



run-hello: build/hello
	./build/hello

run-blinky: build/blinky
	./build/blinky

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
	@go fmt . ./src/examples/hello
	@go fmt ./src/runtime/*.go

gen-device: gen-device-nrf

gen-device-nrf:
	./gen-device.py lib/nrfx/mdk/ src/device/nrf/
	go fmt ./src/device/nrf


# Build the Go compiler.
build/tgo: *.go
	@mkdir -p build
	go build -o build/tgo -i .

# Build IR with the Go compiler.
build/%.bc: src/examples/% src/examples/%/*.go build/tgo src/runtime/*.go build/runtime-$(TARGET)-combined.bc
	./build/tgo $(TGOFLAGS) -runtime build/runtime-$(TARGET)-combined.bc -o $@ $(subst src/,,$<)

# Compile and optimize bitcode file.
build/%.o: build/%.bc
	$(OPT) -Oz -enable-coroutines -o $< $<
	$(LLC) -filetype=obj -o $@ $<

# Compile C sources for the runtime.
build/%.bc: src/runtime/%.c src/runtime/*.h
	@mkdir -p build
	clang $(CFLAGS) -c -o $@ $<

# Compile LLVM bitcode from LLVM source.
build/%.bc: src/runtime/%.ll
	$(LLAS) -o $@ $<

# Compile system_* file for the nRF.
build/nrfx_%.bc: lib/nrfx/mdk/%.c
	@mkdir -p build
	clang $(CFLAGS) -c -o $@ $^

# Compile startup_* file for the nRF.
build/nrfx_%.o: lib/nrfx/mdk/gcc_%.S
	@mkdir -p build
	clang $(CFLAGS) -D__STARTUP_CLEAR_BSS -c -o $@ $^

build/%.o: %.S
	$(AS) -o $@ $^

# Merge all runtime LLVM files together in a single bitcode file.
build/runtime-$(TARGET)-combined.bc: $(RUNTIME_PARTS)
	$(LINK) -o $@ $^

# Generate output ELF executable.
build/%: build/%.o $(OBJ)
	$(LD) -o $@ $^

# Generate output ELF for use in objcopy (on a microcontroller).
build/%.elf: build/%.o $(OBJ)
	$(LD) -o $@ $^
	$(SIZE) $@

# Convert executable to Intel hex file (for flashing).
build/%.hex: build/%.elf
	$(OBJCOPY) -O ihex $^ $@
