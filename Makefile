
# aliases
all: tinygo

# Default build and source directories, as created by `make llvm-build`.
LLVM_BUILDDIR ?= llvm-build
LLVM_PROJECTDIR ?= llvm-project
CLANG_SRC ?= $(LLVM_PROJECTDIR)/clang
LLD_SRC ?= $(LLVM_PROJECTDIR)/lld

# Try to autodetect LLVM build tools.
detect = $(shell command -v $(1) 2> /dev/null && echo $(1))
CLANG ?= $(word 1,$(abspath $(call detect,llvm-build/bin/clang))$(call detect,clang-11)$(call detect,clang))
LLVM_AR ?= $(word 1,$(abspath $(call detect,llvm-build/bin/llvm-ar))$(call detect,llvm-ar-11)$(call detect,llvm-ar))
LLVM_NM ?= $(word 1,$(abspath $(call detect,llvm-build/bin/llvm-nm))$(call detect,llvm-nm-11)$(call detect,llvm-nm))

# Go binary and GOROOT to select
GO ?= go
export GOROOT = $(shell $(GO) env GOROOT)

# Flags to pass to go test.
GOTESTFLAGS ?= -v

# md5sum binary
MD5SUM = md5sum

# tinygo binary for tests
TINYGO ?= $(word 1,$(call detect,tinygo)$(call detect,build/tinygo))

# Use CCACHE for LLVM if possible
ifneq (, $(shell command -v ccache 2> /dev/null))
    LLVM_OPTION += '-DLLVM_CCACHE_BUILD=ON'
endif

# Allow enabling LLVM assertions
ifeq (1, $(ASSERT))
    LLVM_OPTION += '-DLLVM_ENABLE_ASSERTIONS=ON'
else
    LLVM_OPTION += '-DLLVM_ENABLE_ASSERTIONS=OFF'
endif

.PHONY: all tinygo test $(LLVM_BUILDDIR) llvm-source clean fmt gen-device gen-device-nrf gen-device-nxp gen-device-avr gen-device-rp

LLVM_COMPONENTS = all-targets analysis asmparser asmprinter bitreader bitwriter codegen core coroutines coverage debuginfodwarf debuginfopdb executionengine frontendopenmp instrumentation interpreter ipo irreader libdriver linker lto mc mcjit objcarcopts option profiledata scalaropts support target windowsmanifest

ifeq ($(OS),Windows_NT)
    EXE = .exe
    START_GROUP = -Wl,--start-group
    END_GROUP = -Wl,--end-group

    # LLVM compiled using MinGW on Windows appears to have problems with threads.
    # Without this flag, linking results in errors like these:
    #     libLLVMSupport.a(Threading.cpp.obj):Threading.cpp:(.text+0x55): undefined reference to `std::thread::hardware_concurrency()'
    LLVM_OPTION += -DLLVM_ENABLE_THREADS=OFF -DLLVM_ENABLE_PIC=OFF

    CGO_CPPFLAGS += -DCINDEX_NO_EXPORTS
    CGO_LDFLAGS += -static -static-libgcc -static-libstdc++
    CGO_LDFLAGS_EXTRA += -lversion

    BINARYEN_OPTION += -DCMAKE_EXE_LINKER_FLAGS='-static-libgcc -static-libstdc++'

else ifeq ($(shell uname -s),Darwin)
    MD5SUM = md5
else ifeq ($(shell uname -s),FreeBSD)
    MD5SUM = md5
    START_GROUP = -Wl,--start-group
    END_GROUP = -Wl,--end-group
else
    START_GROUP = -Wl,--start-group
    END_GROUP = -Wl,--end-group
endif

# Libraries that should be linked in for the statically linked Clang.
CLANG_LIB_NAMES = clangAnalysis clangAST clangASTMatchers clangBasic clangCodeGen clangCrossTU clangDriver clangDynamicASTMatchers clangEdit clangFormat clangFrontend clangFrontendTool clangHandleCXX clangHandleLLVM clangIndex clangLex clangParse clangRewrite clangRewriteFrontend clangSema clangSerialization clangTooling clangToolingASTDiff clangToolingCore clangToolingInclusions
CLANG_LIBS = $(START_GROUP) $(addprefix -l,$(CLANG_LIB_NAMES)) $(END_GROUP) -lstdc++

# Libraries that should be linked in for the statically linked LLD.
LLD_LIB_NAMES = lldCOFF lldCommon lldCore lldDriver lldELF lldMachO lldMinGW lldReaderWriter lldWasm lldYAML
LLD_LIBS = $(START_GROUP) $(addprefix -l,$(LLD_LIB_NAMES)) $(END_GROUP)

# Other libraries that are needed to link TinyGo.
EXTRA_LIB_NAMES = LLVMInterpreter

# All libraries to be built and linked with the tinygo binary (lib/lib*.a).
LIB_NAMES = clang $(CLANG_LIB_NAMES) $(LLD_LIB_NAMES) $(EXTRA_LIB_NAMES)

# These build targets appear to be the only ones necessary to build all TinyGo
# dependencies. Only building a subset significantly speeds up rebuilding LLVM.
# The Makefile rules convert a name like lldELF to lib/liblldELF.a to match the
# library path (for ninja).
# This list also includes a few tools that are necessary as part of the full
# TinyGo build.
NINJA_BUILD_TARGETS = clang llvm-config llvm-ar llvm-nm $(addprefix lib/lib,$(addsuffix .a,$(LIB_NAMES)))

# For static linking.
ifneq ("$(wildcard $(LLVM_BUILDDIR)/bin/llvm-config*)","")
    CGO_CPPFLAGS+=$(shell $(LLVM_BUILDDIR)/bin/llvm-config --cppflags) -I$(abspath $(LLVM_BUILDDIR))/tools/clang/include -I$(abspath $(CLANG_SRC))/include -I$(abspath $(LLD_SRC))/include
    CGO_CXXFLAGS=-std=c++14
    CGO_LDFLAGS+=-L$(abspath $(LLVM_BUILDDIR)/lib) -lclang $(CLANG_LIBS) $(LLD_LIBS) $(shell $(LLVM_BUILDDIR)/bin/llvm-config --ldflags --libs --system-libs $(LLVM_COMPONENTS)) -lstdc++ $(CGO_LDFLAGS_EXTRA)
endif


clean:
	@rm -rf build

FMT_PATHS = ./*.go builder cgo compiler interp loader src/device/arm src/examples src/machine src/os src/reflect src/runtime src/sync src/syscall src/testing src/internal/reflectlite transform
fmt:
	@gofmt -l -w $(FMT_PATHS)
fmt-check:
	@unformatted=$$(gofmt -l $(FMT_PATHS)); [ -z "$$unformatted" ] && exit 0; echo "Unformatted:"; for fn in $$unformatted; do echo "  $$fn"; done; exit 1


gen-device: gen-device-avr gen-device-esp gen-device-nrf gen-device-sam gen-device-sifive gen-device-kendryte gen-device-nxp gen-device-rp
ifneq ($(STM32), 0)
gen-device: gen-device-stm32
endif

gen-device-avr:
	@if [ ! -e lib/avr/README.md ]; then echo "Submodules have not been downloaded. Please download them using:\n  git submodule update --init"; exit 1; fi
	$(GO) build -o ./build/gen-device-avr ./tools/gen-device-avr/
	./build/gen-device-avr lib/avr/packs/atmega src/device/avr/
	./build/gen-device-avr lib/avr/packs/tiny src/device/avr/
	@GO111MODULE=off $(GO) fmt ./src/device/avr

build/gen-device-svd: ./tools/gen-device-svd/*.go
	$(GO) build -o $@ ./tools/gen-device-svd/

gen-device-esp: build/gen-device-svd
	./build/gen-device-svd -source=https://github.com/posborne/cmsis-svd/tree/master/data/Espressif-Community -interrupts=software lib/cmsis-svd/data/Espressif-Community/ src/device/esp/
	./build/gen-device-svd -source=https://github.com/posborne/cmsis-svd/tree/master/data/Espressif -interrupts=software lib/cmsis-svd/data/Espressif/ src/device/esp/
	GO111MODULE=off $(GO) fmt ./src/device/esp

gen-device-nrf: build/gen-device-svd
	./build/gen-device-svd -source=https://github.com/NordicSemiconductor/nrfx/tree/master/mdk lib/nrfx/mdk/ src/device/nrf/
	GO111MODULE=off $(GO) fmt ./src/device/nrf

gen-device-nxp: build/gen-device-svd
	./build/gen-device-svd -source=https://github.com/posborne/cmsis-svd/tree/master/data/NXP lib/cmsis-svd/data/NXP/ src/device/nxp/
	GO111MODULE=off $(GO) fmt ./src/device/nxp

gen-device-sam: build/gen-device-svd
	./build/gen-device-svd -source=https://github.com/posborne/cmsis-svd/tree/master/data/Atmel lib/cmsis-svd/data/Atmel/ src/device/sam/
	GO111MODULE=off $(GO) fmt ./src/device/sam

gen-device-sifive: build/gen-device-svd
	./build/gen-device-svd -source=https://github.com/posborne/cmsis-svd/tree/master/data/SiFive-Community -interrupts=software lib/cmsis-svd/data/SiFive-Community/ src/device/sifive/
	GO111MODULE=off $(GO) fmt ./src/device/sifive

gen-device-kendryte: build/gen-device-svd
	./build/gen-device-svd -source=https://github.com/posborne/cmsis-svd/tree/master/data/Kendryte-Community -interrupts=software lib/cmsis-svd/data/Kendryte-Community/ src/device/kendryte/
	GO111MODULE=off $(GO) fmt ./src/device/kendryte

gen-device-stm32: build/gen-device-svd
	./build/gen-device-svd -source=https://github.com/tinygo-org/stm32-svd lib/stm32-svd/svd src/device/stm32/
	GO111MODULE=off $(GO) fmt ./src/device/stm32

gen-device-rp: build/gen-device-svd
	./build/gen-device-svd -source=https://github.com/posborne/cmsis-svd/tree/master/data/RaspberryPi lib/cmsis-svd/data/RaspberryPi/ src/device/rp/
	GO111MODULE=off $(GO) fmt ./src/device/rp

# Get LLVM sources.
$(LLVM_PROJECTDIR)/llvm:
	git clone -b xtensa_release_12.0.1 --depth=1 https://github.com/tinygo-org/llvm-project $(LLVM_PROJECTDIR)
llvm-source: $(LLVM_PROJECTDIR)/llvm

# Configure LLVM.
TINYGO_SOURCE_DIR=$(shell pwd)
$(LLVM_BUILDDIR)/build.ninja: llvm-source
	mkdir -p $(LLVM_BUILDDIR) && cd $(LLVM_BUILDDIR) && cmake -G Ninja $(TINYGO_SOURCE_DIR)/$(LLVM_PROJECTDIR)/llvm "-DLLVM_TARGETS_TO_BUILD=X86;ARM;AArch64;RISCV;WebAssembly" "-DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD=AVR;Xtensa" -DCMAKE_BUILD_TYPE=Release -DLIBCLANG_BUILD_STATIC=ON -DLLVM_ENABLE_TERMINFO=OFF -DLLVM_ENABLE_ZLIB=OFF -DLLVM_ENABLE_LIBEDIT=OFF -DLLVM_ENABLE_Z3_SOLVER=OFF -DLLVM_ENABLE_OCAMLDOC=OFF -DLLVM_ENABLE_LIBXML2=OFF -DLLVM_ENABLE_PROJECTS="clang;lld" -DLLVM_TOOL_CLANG_TOOLS_EXTRA_BUILD=OFF -DCLANG_ENABLE_STATIC_ANALYZER=OFF -DCLANG_ENABLE_ARCMT=OFF $(LLVM_OPTION)

# Build LLVM.
$(LLVM_BUILDDIR): $(LLVM_BUILDDIR)/build.ninja
	cd $(LLVM_BUILDDIR) && ninja $(NINJA_BUILD_TARGETS)

# Build Binaryen
.PHONY: binaryen
binaryen: build/wasm-opt$(EXE)
build/wasm-opt$(EXE):
	mkdir -p build
	cd lib/binaryen && cmake -G Ninja . -DBUILD_STATIC_LIB=ON $(BINARYEN_OPTION) && ninja bin/wasm-opt$(EXE)
	cp lib/binaryen/bin/wasm-opt$(EXE) build/wasm-opt$(EXE)

# Build wasi-libc sysroot
.PHONY: wasi-libc
wasi-libc: lib/wasi-libc/sysroot/lib/wasm32-wasi/libc.a
lib/wasi-libc/sysroot/lib/wasm32-wasi/libc.a:
	@if [ ! -e lib/wasi-libc/Makefile ]; then echo "Submodules have not been downloaded. Please download them using:\n  git submodule update --init"; exit 1; fi
	cd lib/wasi-libc && make -j4 WASM_CFLAGS="-O2 -g -DNDEBUG" WASM_CC=$(CLANG) WASM_AR=$(LLVM_AR) WASM_NM=$(LLVM_NM)


# Build the Go compiler.
tinygo:
	@if [ ! -f "$(LLVM_BUILDDIR)/bin/llvm-config" ]; then echo "Fetch and build LLVM first by running:"; echo "  make llvm-source"; echo "  make $(LLVM_BUILDDIR)"; exit 1; fi
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) build -buildmode exe -o build/tinygo$(EXE) -tags byollvm -ldflags="-X main.gitSha1=`git rev-parse --short HEAD`" .

test: wasi-libc
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=20m -buildmode exe -tags byollvm ./builder ./cgo ./compileopts ./compiler ./interp ./transform .

TEST_PACKAGES = \
	compress/bzip2 \
	compress/flate \
	compress/zlib \
	container/heap \
	container/list \
	container/ring \
	crypto/des \
	crypto/dsa \
	crypto/elliptic/internal/fiat \
	crypto/internal/subtle \
	crypto/md5 \
	crypto/rc4 \
	crypto/sha1 \
	crypto/sha256 \
	crypto/sha512 \
	debug/macho \
	encoding \
	encoding/ascii85 \
	encoding/base32 \
	encoding/csv \
	encoding/hex \
	go/scanner \
	hash \
	hash/adler32 \
	hash/crc64 \
	hash/fnv \
	html \
	index/suffixarray \
	internal/itoa \
	internal/profile \
	math \
	math/cmplx \
	net/http/internal/ascii \
	net/mail \
	os \
	path \
	reflect \
	testing \
	testing/iotest \
	text/scanner \
	unicode \
	unicode/utf16 \
	unicode/utf8 \

# Test known-working standard library packages.
# TODO: parallelize, and only show failing tests (no implied -v flag).
.PHONY: tinygo-test
tinygo-test:
	$(TINYGO) test $(TEST_PACKAGES)

build/smoketest:
	mkdir build/smoketest

.PHONY: _smoketest
_smoketest:

define smokeTarg
$(subst -,_,$(1))_ext = $(2)

build/smoketest/$(1): build/smoketest
	if [ ! -d $$@ ]; then mkdir $$@; fi

build/smoketest/$(1)/%.$(2): build/smoketest/$(1) _smoketest
	$(TINYGO) build -size short -o $$@ -target=$(1) examples/$$* > build/smoketest/$(1)/$$*.txt
endef

# Configure file extensions for smoke tests.
hexmcus = arduino arduino-mega1280 arduino-nano arduino-nano33 nano-33-ble arduino-mkrwifi1010 \
	pca10031 pca10040 pca10040-s132v6 pca10056 pca10056-s140v7 pca10059 \
	microbit microbit-s110v8 microbit-v2 microbit-v2-s113v7 \
	nrf52840-mdk \
	itsybitsy-m0 itsybitsy-m4 itsybitsy-nrf52840 \
	feather-m0 feather-m4 feather-m4-can feather-stm32f405 feather-nrf52840 feather-nrf52840-sense \
	trinket-m0 \
	grandcentral-m4 metro-m4-airlift \
	circuitplay-express circuitplay-bluefruit bluepill \
	pybadge pyportal pygamer qtpy \
	teensy36 teensy40 \
	pinetime-devkit0 x9pro \
	pico nano-rp2040 feather-rp2040 \
	nucleo-f103rb nucleo-f722ze nucleo-l031k6 nucleo-l432kc nucleo-l552ze nucleo-wl55jc \
	stm32f4disco stm32f4disco-1 \
	hifive1b hifive1-qemu maixbit \
	reelboard reelboard-s140v7 \
	particle-argon particle-boron particle-xenon \
	atmega1284p \
	clue-alpha wioterminal xiao p1am-100 atsame54-xpro lgt92 lorae5 digispark
$(foreach mcu,$(hexmcus),$(eval $(call smokeTarg,$(mcu),hex)))

binmcus = esp32-mini32 nodemcu m5stack m5stack-core2 esp32c3
$(foreach mcu,$(binmcus),$(eval $(call smokeTarg,$(mcu),bin)))

$(eval $(call smokeTarg,gameboy-advance,gba))

$(eval $(call smokeTarg,nintendoswitch,nro))

# Set default smoke tests for microcontroller boards/devices/etc.

# Generic
pca10031_smoke = blinky1
pca10040_smoke = blinky1 adc blinkm blinky2 button button2 echo mcp3008 memstats pininterrupt serial systick test
pca10040_s132v6_smoke = blinky1
pca10056_smoke = blinky1 blinky2
pca10056_s140v7_smoke = blinky1
pca10059_smoke = blinky1 blinky2
circuitplay_express_smoke = blinky1 i2s dac
circuitplay_bluefruit_smoke = blinky1
microbit_smoke = microbit-blink echo
microbit_s110v8_smoke = echo
microbit_v2_smoke = microbit-blink
microbit_v2_s113v7_smoke = microbit-blink
nrf52840_mdk_smoke = blinky1
reelboard_smoke = blinky1 blinky2
reelboard_s140v7_smoke = blinky1
itsybitsy_m0_smoke = blinky1 pwm
itsybitsy_m4_smoke = blinky1 pwm
itsybitsy_nrf52840_smoke = blinky1
feather_m0_smoke = blinky1
feather_m4_smoke = blinky1 pwm
feather_m4_can_smoke = blinky1 can
feather_rp2040_smoke = blinky1
feather_nrf52840_smoke = blinky1
feather_nrf52840_sense_smoke = blinky1
trinket_m0_smoke = blinky1
grandcentral_m4_smoke = blinky1
clue_alpha_smoke = blinky1
gameboy_advance_smoke = gba-display
pybadge_smoke = blinky1
metro_m4_airlift_smoke = blinky1
pyportal_smoke = blinky1 dac
particle_argon_smoke = blinky1
particle_boron_smoke = blinky1
particle_xenon_smoke = blinky1
pinetime_devkit0_smoke = blinky1
x9pro_smoke = blinky1
wioterminal_smoke = blinky1
pygamer_smoke = blinky1
xiao_smoke = blinky1
qtpy_smoke = serial
teensy36_smoke = blinky1
teensy40_smoke = blinky1
p1am_100_smoke = blinky1
atsame54_xpro_smoke = blinky1 can
arduino_nano33_smoke = blinky1
arduino_mkrwifi1010_smoke = blinky1
pico_smoke = blinky1
nano_33_ble_smoke = blinky1
nano_rp2040_smoke = blinky1
nintendoswitch_smoke = serial

# STM32
bluepill_smoke = blinky1
feather_stm32f405_smoke = blinky1
lgt92_smoke = blinky1
nucleo_f103rb_smoke = blinky1
nucleo_f722ze_smoke = blinky1
nucleo_l031k6_smoke = blinky1
nucleo_l432kc_smoke = blinky1
nucleo_l552ze_smoke = blinky1
nucleo_wl55jc_smoke = blinky1
stm32f4disco_smoke = blinky1 blinky2
stm32f4disco_1_smoke = blinky1 pwm
lorae5_smoke = blinky1

# AVR
atmega1284p_smoke = serial
arduino_smoke = blinky1 pwm
arduino_mega1280_smoke = blinky1 pwm
arduino_nano_smoke = blinky1
digispark_smoke = blinky1

# XTENSA
esp32_mini32_smoke = blinky1
nodemcu_smoke = blinky1
m5stack_smoke = serial
m5stack_core2_smoke = serial

genericTargets = pca10031 pca10040 pca10040-s132v6 pca10056 pca10056-s140v7 pca10059 \
	circuitplay-express circuitplay-bluepill \
	microbit microbit-s110v8 microbit-v2 microbit-v2-s113v7 \
	nrf52840-mdk \
	reelboard reelboard-s140v7 \
	itsybitsy-m0 itsybitsy-m4 itsybitsy-nrf52840 \
	feather-m0 feather-m4 feather-m4-can feather-rp2040 feather-nrf52840 feather-nrf52840-sense \
	trinket-m0 \
	grandcentral-m4 \
	metro-m4-airlift \
	clue-alpha \
	gameboy-advance nintendoswitch \
	pybadge pyportal pygamer \
	particle-argon particle-boron particle-xenon \
	pinetime-devkit0 x9pro \
	wioterminal \
	xiao \
	qtpy \
	teensy36 teensy40 \
	p1am-100 \
	atsame54-xpro \
	arduino-nano33 \
	arduino-mkrwifi1010 \
	pico \
	nano-33-ble \
	nano-rp2040

stm32Targets = bluepill \
	feather-stm32f405 \
	lgt92 \
	nucleo-f103rb nucleo-f722ze nucleo-l031k6 nucleo-l432kc nucleo-l552ze nucleo-wl55jc \
	stm32f4disco stm32f4disco-1 \
	lorae5

avrTargets = atmega1284p \
	arduino \
	arduino-mega1280 \
	arduino-nano \
	digispark

xtensaTargets = esp32-mini32 \
	nodemcu \
	m5stack m5stack-core2

targets = $(genericTargets)

ifneq ($(STM32), 0)
targets += $(stm32Targets)
endif
ifneq ($(AVR), 0)
targets += $(avrTargets)
endif
ifneq ($(XTENSA), 0)
targets += $(xtensaTargets)
endif

define smoke
$(foreach example,$($(subst -,_,$(1))_smoke),build/smoketest/$(1)/$(example).$($(subst -,_,$(1))_ext))
endef
smoketest: $(foreach targ,$(targets),$(call smoke,$(targ))) 

# Set up webassembly stuff.
build/smoketest/fake build/smoketest/wasm: build/smoketest
	if [ ! -d $@ ]; then mkdir $@; fi

build/smoketest/wasm/%.wasm: build/smoketest/wasm _smoketest
	$(TINYGO) build -size short -o $@ -target=wasm examples/wasm/$* > build/smoketest/wasm/$*.txt

define fakeSmoke
build/smoketest/fake/$(1): build/smoketest/fake
	if [ ! -d $$@ ]; then mkdir $$@; fi

build/smoketest/fake/$(1)/%.wasm: build/smoketest/fake/$(1) _smoketest
	$(TINYGO) build -size short -o $$@ -tags=$(subst -,_,$(1)) examples/$$* > build/smoketest/fake/$(1)/$$*.txt
endef

fakeTargets = arduino hifive1b reelboard pca10040 pca10056 circuitplay-express

$(foreach targ,$(fakeTargets),$(eval $(call fakeSmoke,$(targ))))

ifneq ($(WASM), 0)
smoketest: \
	build/smoketest/wasm/export.wasm \
	build/smoketest/wasm/main.wasm \
	build/smoketest/fake/arduino/blinky1.wasm \
	build/smoketest/fake/hifive1b/blinky1.wasm \
	build/smoketest/fake/reelboard/blinky1.wasm \
	build/smoketest/fake/pca10040/blinky2.wasm \
	build/smoketest/fake/pca10056/blinky2.wasm \
	build/smoketest/fake/circuitplay-express/blinky1.wasm
endif

# Test compiler flags.
build/smoketest/pca10040/blinky1-nogc-nosched.hex: build/smoketest/pca10040 _smoketest
	$(TINYGO) build -size short -o $@ -target=pca10040 -gc=none -scheduler=none examples/blinky1 > build/smoketest/pca10040/blinky1-nogc-nosched.txt

build/smoketest/pca10040/blinky1-opt1.hex: build/smoketest/pca10040 _smoketest
	$(TINYGO) build -size short -o $@ -target=pca10040 -opt=1 examples/blinky1 > build/smoketest/pca10040/blinky1-opt1.txt

build/smoketest/pca10040/echo-noserial.hex: build/smoketest/pca10040 _smoketest
	$(TINYGO) build -size short -o $@ -target=pca10040 -serial=none examples/echo > build/smoketest/pca10040/echo-noserial.txt

build/smoketest/pca10040/stdlib-opt0.hex: build/smoketest/pca10040 _smoketest
	$(TINYGO) build -size short -o $@ -target=pca10040 -opt=0 ./testdata/stdlib.go > build/smoketest/pca10040/stdlib-opt0.txt

build/smoketest/arduino/blinky1-tasks.hex: build/smoketest/arduino _smoketest
	$(TINYGO) build -size short -o $@ -target=arduino -scheduler=tasks examples/blinky1 > build/smoketest/arduino/blinky1-tasks.txt

build/smoketest/digispark/blinky1-leaking.hex: build/smoketest/digispark _smoketest
	$(TINYGO) build -size short -o $@ -target=digispark -gc=leaking examples/blinky1 > build/smoketest/digispark/blinky1-leaking.txt

build/smoketest/misc: build/smoketest
	if [ ! -d $@ ]; then mkdir $@; fi

build/smoketest/misc/cgo-linux-arm.elf: build/smoketest/misc _smoketest
	GOOS=linux GOARCH=arm $(TINYGO) build -size short -o $@ ./testdata/cgo > build/smoketest/misc/cgo-linux-arm.txt

build/smoketest/misc/cgo-windows-amd64.exe: build/smoketest/misc _smoketest
	GOOS=windows GOARCH=amd64 $(TINYGO) build -size short -o $@ ./testdata/cgo > build/smoketest/misc/cgo-windows-amd64.txt

build/smoketest/misc/serial-leaking-nosched.elf: build/smoketest/misc _smoketest
	$(TINYGO) build -size short -gc=leaking -scheduler=none -o $@ examples/serial > build/smoketest/misc/serial-leaking-nosched.txt

smoketest: \
	build/smoketest/pca10040/blinky1-nogc-nosched.hex \
	build/smoketest/pca10040/blinky1-opt1.hex \
	build/smoketest/pca10040/echo-noserial.hex \
	build/smoketest/pca10040/stdlib-opt0.hex \
	build/smoketest/misc/cgo-linux-arm.elf \
	build/smoketest/misc/cgo-windows-amd64.exe
ifneq ($(AVR), 0)
smoketest: \
	build/smoketest/arduino/blinky1-tasks.hex \
	build/smoketest/digispark/blinky1-leaking.hex
endif
ifneq ($(OS),Windows_NT)
smoketest: \
	build/smoketest/misc/serial-leaking-nosched.elf
endif

# Print the smoketest results.
smoketest:
	@for i in $(sort $^); do \
		$(MD5SUM) $$i; \
		cat $${i%.*}.txt; \
	done

wasmtest:
	$(GO) test ./tests/wasm

build/release: tinygo gen-device wasi-libc binaryen
	@mkdir -p build/release/tinygo/bin
	@mkdir -p build/release/tinygo/lib/clang/include
	@mkdir -p build/release/tinygo/lib/CMSIS/CMSIS
	@mkdir -p build/release/tinygo/lib/compiler-rt/lib
	@mkdir -p build/release/tinygo/lib/mingw-w64/mingw-w64-crt/lib-common
	@mkdir -p build/release/tinygo/lib/mingw-w64/mingw-w64-headers/defaults
	@mkdir -p build/release/tinygo/lib/musl/arch
	@mkdir -p build/release/tinygo/lib/musl/crt
	@mkdir -p build/release/tinygo/lib/musl/src
	@mkdir -p build/release/tinygo/lib/nrfx
	@mkdir -p build/release/tinygo/lib/picolibc/newlib/libc
	@mkdir -p build/release/tinygo/lib/picolibc/newlib/libm
	@mkdir -p build/release/tinygo/lib/wasi-libc
	@mkdir -p build/release/tinygo/pkg/armv6m-unknown-unknown-eabi
	@mkdir -p build/release/tinygo/pkg/armv7m-unknown-unknown-eabi
	@mkdir -p build/release/tinygo/pkg/armv7em-unknown-unknown-eabi
	@echo copying source files
	@cp -p  build/tinygo$(EXE)           build/release/tinygo/bin
	@cp -p  build/wasm-opt$(EXE)         build/release/tinygo/bin
	@cp -p $(abspath $(CLANG_SRC))/lib/Headers/*.h build/release/tinygo/lib/clang/include
	@cp -rp lib/CMSIS/CMSIS/Include      build/release/tinygo/lib/CMSIS/CMSIS
	@cp -rp lib/CMSIS/README.md          build/release/tinygo/lib/CMSIS
	@cp -rp lib/compiler-rt/lib/builtins build/release/tinygo/lib/compiler-rt/lib
	@cp -rp lib/compiler-rt/LICENSE.TXT  build/release/tinygo/lib/compiler-rt
	@cp -rp lib/compiler-rt/README.txt   build/release/tinygo/lib/compiler-rt
	@cp -rp lib/musl/arch/aarch64        build/release/tinygo/lib/musl/arch
	@cp -rp lib/musl/arch/arm            build/release/tinygo/lib/musl/arch
	@cp -rp lib/musl/arch/generic        build/release/tinygo/lib/musl/arch
	@cp -rp lib/musl/arch/i386           build/release/tinygo/lib/musl/arch
	@cp -rp lib/musl/arch/x86_64         build/release/tinygo/lib/musl/arch
	@cp -rp lib/musl/crt/crt1.c          build/release/tinygo/lib/musl/crt
	@cp -rp lib/musl/COPYRIGHT           build/release/tinygo/lib/musl
	@cp -rp lib/musl/include             build/release/tinygo/lib/musl
	@cp -rp lib/musl/src/env             build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/errno           build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/exit            build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/include         build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/internal        build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/malloc          build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/mman            build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/signal          build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/stdio           build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/string          build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/thread          build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/time            build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/unistd          build/release/tinygo/lib/musl/src
	@cp -rp lib/mingw-w64/mingw-w64-crt/def-include                 build/release/tinygo/lib/mingw-w64/mingw-w64-crt
	@cp -rp lib/mingw-w64/mingw-w64-crt/lib-common/api-ms-win-crt-* build/release/tinygo/lib/mingw-w64/mingw-w64-crt/lib-common
	@cp -rp lib/mingw-w64/mingw-w64-crt/lib-common/kernel32.def.in  build/release/tinygo/lib/mingw-w64/mingw-w64-crt/lib-common
	@cp -rp lib/mingw-w64/mingw-w64-headers/crt/                    build/release/tinygo/lib/mingw-w64/mingw-w64-headers
	@cp -rp lib/mingw-w64/mingw-w64-headers/defaults/include        build/release/tinygo/lib/mingw-w64/mingw-w64-headers/defaults
	@cp -rp lib/nrfx/*                   build/release/tinygo/lib/nrfx
	@cp -rp lib/picolibc/newlib/libc/ctype       build/release/tinygo/lib/picolibc/newlib/libc
	@cp -rp lib/picolibc/newlib/libc/include     build/release/tinygo/lib/picolibc/newlib/libc
	@cp -rp lib/picolibc/newlib/libc/locale      build/release/tinygo/lib/picolibc/newlib/libc
	@cp -rp lib/picolibc/newlib/libc/string      build/release/tinygo/lib/picolibc/newlib/libc
	@cp -rp lib/picolibc/newlib/libc/tinystdio   build/release/tinygo/lib/picolibc/newlib/libc
	@cp -rp lib/picolibc/newlib/libm/common      build/release/tinygo/lib/picolibc/newlib/libm
	@cp -rp lib/picolibc-stdio.c         build/release/tinygo/lib
	@cp -rp lib/wasi-libc/sysroot        build/release/tinygo/lib/wasi-libc/sysroot
	@cp -rp src                          build/release/tinygo/src
	@cp -rp targets                      build/release/tinygo/targets
	./build/tinygo build-library -target=armv6m-unknown-unknown-eabi  -o build/release/tinygo/pkg/armv6m-unknown-unknown-eabi/compiler-rt compiler-rt
	./build/tinygo build-library -target=armv7m-unknown-unknown-eabi  -o build/release/tinygo/pkg/armv7m-unknown-unknown-eabi/compiler-rt compiler-rt
	./build/tinygo build-library -target=armv7em-unknown-unknown-eabi -o build/release/tinygo/pkg/armv7em-unknown-unknown-eabi/compiler-rt compiler-rt
	./build/tinygo build-library -target=armv6m-unknown-unknown-eabi  -o build/release/tinygo/pkg/armv6m-unknown-unknown-eabi/picolibc picolibc
	./build/tinygo build-library -target=armv7m-unknown-unknown-eabi  -o build/release/tinygo/pkg/armv7m-unknown-unknown-eabi/picolibc picolibc
	./build/tinygo build-library -target=armv7em-unknown-unknown-eabi -o build/release/tinygo/pkg/armv7em-unknown-unknown-eabi/picolibc picolibc

release: build/release
	tar -czf build/release.tar.gz -C build/release tinygo

deb: build/release
	@mkdir -p build/release-deb/usr/local/bin
	@mkdir -p build/release-deb/usr/local/lib
	cp -ar build/release/tinygo build/release-deb/usr/local/lib/tinygo
	ln -sf ../lib/tinygo/bin/tinygo build/release-deb/usr/local/bin/tinygo
	fpm -f -s dir -t deb -n tinygo -v $(shell grep "const Version = " goenv/version.go | awk '{print $$NF}') -m '@tinygo-org' --description='TinyGo is a Go compiler for small places.' --license='BSD 3-Clause' --url=https://tinygo.org/ --deb-changelog CHANGELOG.md -p build/release.deb -C ./build/release-deb
