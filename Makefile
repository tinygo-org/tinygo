
# aliases
all: tgo
tgo: build/tgo

.PHONY: all tgo run-hello run-blinky flash-blinky clean

# Custom LLVM toolchain.
LLVM =
LINK = $(LLVM)llvm-link
LLC = $(LLVM)llc

CFLAGS = -Wall -Werror -Os -g -fno-exceptions -flto -ffunction-sections -fdata-sections $(LLFLAGS)

RUNTIME = build/runtime.bc

ifeq ($(TARGET),pca10040)
GCC = arm-none-eabi-gcc
LD = arm-none-eabi-ld -T arm.ld
OBJCOPY = arm-none-eabi-objcopy
LLFLAGS += -target armv7m-none-eabi
TGOFLAGS += -target $(TARGET)
CFLAGS += -I$(CURDIR)/src/runtime
CFLAGS += -I$(CURDIR)/lib/nrfx
CFLAGS += -I$(CURDIR)/lib/nrfx/mdk
CFLAGS += -I$(CURDIR)/lib/CMSIS/CMSIS/Include
CFLAGS += -DNRF52832_XXAA
CFLAGS += -Wno-uninitialized
RUNTIME += build/runtime_nrf.bc
RUNTIME += build/system_nrf52.bc
OBJ += build/startup_nrf51.o # TODO nrf52, see https://bugs.llvm.org/show_bug.cgi?id=31601

else
# Regular *nix system.
GCC = gcc
LD = clang
endif



run-hello: build/hello
	./build/hello

run-blinky: build/blinky
	./build/blinky

flash-blinky: build/blinky.hex
	nrfjprog -f nrf52 --sectorerase --program $< --reset

clean:
	@rm -rf build



# Build the Go compiler.
build/tgo: *.go
	@mkdir -p build
	go build -o build/tgo -i .

# Build textual IR with the Go compiler.
build/%.ll: src/examples/% src/examples/%/*.go build/tgo src/runtime/*.go
	./build/tgo $(TGOFLAGS) -printir -o $@ $(subst src/,,$<)

# Compile C sources for the runtime.
build/%.bc: src/runtime/%.c src/runtime/*.h
	@mkdir -p build
	clang $(CFLAGS) -c -o $@ $<

# Compile system_* file for the nRF.
build/%.bc: lib/nrfx/mdk/%.c
	@mkdir -p build
	clang $(CFLAGS) -c -o $@ $^

# Compile startup_* file for the nRF.
build/%.o: lib/nrfx/mdk/gcc_%.S
	@mkdir -p build
	clang $(CFLAGS) -c -o $@ $^

# Merge all LLVM files together in a single bitcode file.
build/%.bc: $(RUNTIME) build/%.ll
	$(LINK) -o $@ $^

# Generate an ELF object file from a LLVM bitcode file.
build/%.o: build/%.bc
	$(LLC) -filetype=obj -O2 -o $@ $^

# Generate output ELF executable.
build/%: build/%.o $(OBJ)
	$(LD) -o $@ $^

# Generate output ELF for use in objcopy (on a microcontroller).
build/%.elf: build/%.o $(OBJ)
	$(LD) -o $@ $^

# Convert executable to Intel hex file (for flashing).
build/%.hex: build/%.elf
	$(OBJCOPY) -O ihex $^ $@
