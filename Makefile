
# aliases
all: tinygo

# Default build and source directories, as created by `make llvm-build`.
LLVM_BUILDDIR ?= llvm-build
LLVM_PROJECTDIR ?= llvm-project
CLANG_SRC ?= $(LLVM_PROJECTDIR)/clang
LLD_SRC ?= $(LLVM_PROJECTDIR)/lld

# Try to autodetect LLVM build tools.
# Versions are listed here in descending priority order.
LLVM_VERSIONS = 13 12 11
errifempty = $(if $(1),$(1),$(error $(2)))
detect = $(shell which $(call errifempty,$(firstword $(foreach p,$(2),$(shell command -v $(p) 2> /dev/null && echo $(p)))),failed to locate $(1) at any of: $(2)))
toolSearchPathsVersion = $(1)-$(2)
ifeq ($(shell uname -s),Darwin)
	# Also explicitly search Brew's copy, which is not in PATH by default.
	BREW_PREFIX := $(shell brew --prefix)
	toolSearchPathsVersion += $(BREW_PREFIX)/opt/llvm@$(2)/bin/$(1)-$(2) $(BREW_PREFIX)/opt/llvm@$(2)/bin/$(1)
endif
# First search for a custom built copy, then move on to explicitly version-tagged binaries, then just see if the tool is in path with its normal name.
findLLVMTool = $(call detect,$(1),$(abspath llvm-build/bin/$(1)) $(foreach ver,$(LLVM_VERSIONS),$(call toolSearchPathsVersion,$(1),$(ver))) $(1))
CLANG ?= $(call findLLVMTool,clang)
LLVM_AR ?= $(call findLLVMTool,llvm-ar)
LLVM_NM ?= $(call findLLVMTool,llvm-nm)

# Go binary and GOROOT to select
GO ?= go
export GOROOT = $(shell $(GO) env GOROOT)

# Flags to pass to go test.
GOTESTFLAGS ?= -v

# tinygo binary for tests
TINYGO ?= $(call detect,tinygo,tinygo $(CURDIR)/build/tinygo)

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

    USE_SYSTEM_BINARYEN ?= 1

else ifeq ($(shell uname -s),Darwin)
    CGO_LDFLAGS += -lxar

    USE_SYSTEM_BINARYEN ?= 1

else ifeq ($(shell uname -s),FreeBSD)
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
LLD_LIB_NAMES = lldCOFF lldCommon lldCore lldDriver lldELF lldMachO2 lldMinGW lldReaderWriter lldWasm lldYAML
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
	git clone -b xtensa_release_13.0.0 --depth=1 https://github.com/tinygo-org/llvm-project $(LLVM_PROJECTDIR)
llvm-source: $(LLVM_PROJECTDIR)/llvm

# Configure LLVM.
TINYGO_SOURCE_DIR=$(shell pwd)
$(LLVM_BUILDDIR)/build.ninja: llvm-source
	mkdir -p $(LLVM_BUILDDIR) && cd $(LLVM_BUILDDIR) && cmake -G Ninja $(TINYGO_SOURCE_DIR)/$(LLVM_PROJECTDIR)/llvm "-DLLVM_TARGETS_TO_BUILD=X86;ARM;AArch64;RISCV;WebAssembly" "-DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD=AVR;Xtensa" -DCMAKE_BUILD_TYPE=Release -DLIBCLANG_BUILD_STATIC=ON -DLLVM_ENABLE_TERMINFO=OFF -DLLVM_ENABLE_ZLIB=OFF -DLLVM_ENABLE_LIBEDIT=OFF -DLLVM_ENABLE_Z3_SOLVER=OFF -DLLVM_ENABLE_OCAMLDOC=OFF -DLLVM_ENABLE_LIBXML2=OFF -DLLVM_ENABLE_PROJECTS="clang;lld" -DLLVM_TOOL_CLANG_TOOLS_EXTRA_BUILD=OFF -DCLANG_ENABLE_STATIC_ANALYZER=OFF -DCLANG_ENABLE_ARCMT=OFF $(LLVM_OPTION)

# Build LLVM.
$(LLVM_BUILDDIR): $(LLVM_BUILDDIR)/build.ninja
	cd $(LLVM_BUILDDIR) && ninja $(NINJA_BUILD_TARGETS)

ifneq ($(USE_SYSTEM_BINARYEN),1)
# Build Binaryen
.PHONY: binaryen
binaryen: build/wasm-opt$(EXE)
build/wasm-opt$(EXE):
	mkdir -p build
	cd lib/binaryen && cmake -G Ninja . -DBUILD_STATIC_LIB=ON $(BINARYEN_OPTION) && ninja bin/wasm-opt$(EXE)
	cp lib/binaryen/bin/wasm-opt$(EXE) build/wasm-opt$(EXE)
endif

# Build wasi-libc sysroot
.PHONY: wasi-libc
wasi-libc: lib/wasi-libc/sysroot/lib/wasm32-wasi/libc.a
lib/wasi-libc/sysroot/lib/wasm32-wasi/libc.a:
	@if [ ! -e lib/wasi-libc/Makefile ]; then echo "Submodules have not been downloaded. Please download them using:\n  git submodule update --init"; exit 1; fi
	cd lib/wasi-libc && make -j4 WASM_CFLAGS="-O2 -g -DNDEBUG" MALLOC_IMPL=none WASM_CC=$(CLANG) WASM_AR=$(LLVM_AR) WASM_NM=$(LLVM_NM)


# Build the Go compiler.
tinygo:
	@if [ ! -f "$(LLVM_BUILDDIR)/bin/llvm-config" ]; then echo "Fetch and build LLVM first by running:"; echo "  make llvm-source"; echo "  make $(LLVM_BUILDDIR)"; exit 1; fi
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) build -buildmode exe -o build/tinygo$(EXE) -tags byollvm -ldflags="-X main.gitSha1=`git rev-parse --short HEAD`" .

test: wasi-libc
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=20m -buildmode exe -tags byollvm ./builder ./cgo ./compileopts ./compiler ./interp ./transform .

# Tests that take over a minute in wasi
TEST_PACKAGES_SLOW = \
	compress/bzip2 \
	compress/flate \
	crypto/dsa \
	index/suffixarray \

TEST_PACKAGES_BASE = \
	compress/zlib \
	container/heap \
	container/list \
	container/ring \
	crypto/des \
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
	internal/itoa \
	internal/profile \
	math \
	math/cmplx \
	net/http/internal/ascii \
	net/mail \
	os \
	path \
	reflect \
	sync \
	testing \
	testing/iotest \
	text/scanner \
	unicode \
	unicode/utf16 \
	unicode/utf8 \

# Standard library packages that pass tests natively
TEST_PACKAGES := \
	$(TEST_PACKAGES_BASE)

# archive/zip requires ReadAt, which is not yet supported on windows
ifneq ($(OS),Windows_NT)
TEST_PACKAGES := \
	$(TEST_PACKAGES) \
	archive/zip
endif

# Standard library packages that pass tests on wasi
TEST_PACKAGES_WASI = \
	$(TEST_PACKAGES_BASE)

# Test known-working standard library packages.
# TODO: parallelize, and only show failing tests (no implied -v flag).
.PHONY: tinygo-test
tinygo-test:
	$(TINYGO) test $(TEST_PACKAGES) $(TEST_PACKAGES_SLOW)
tinygo-test-fast:
	$(TINYGO) test $(TEST_PACKAGES)
tinygo-bench:
	$(TINYGO) test -bench . $(TEST_PACKAGES) $(TEST_PACKAGES_SLOW)
tinygo-bench-fast:
	$(TINYGO) test -bench . $(TEST_PACKAGES)
tinygo-test-wasi:
	$(TINYGO) test -target wasi $(TEST_PACKAGES_WASI) $(TEST_PACKAGES_SLOW)
tinygo-test-wasi-fast:
	$(TINYGO) test -target wasi $(TEST_PACKAGES_WASI)
tinygo-bench-wasi:
	$(TINYGO) test -target wasi -bench . $(TEST_PACKAGES_WASI) $(TEST_PACKAGES_SLOW)
tinygo-bench-wasi-fast:
	$(TINYGO) test -target wasi -bench . $(TEST_PACKAGES_WASI)

# Test external packages in a large corpus.
test-corpus:
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=1h -buildmode exe -tags byollvm -run TestCorpus . -corpus=testdata/corpus.yaml
test-corpus-fast:
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=1h -buildmode exe -tags byollvm -run TestCorpus -short . -corpus=testdata/corpus.yaml
test-corpus-wasi: wasi-libc
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=1h -buildmode exe -tags byollvm -run TestCorpus . -corpus=testdata/corpus.yaml -target=wasi


.PHONY: smoketest smoketest-commands
smoketest:
	$(TINYGO) version
	# regression test for #2563
	cd tests/os/smoke && $(TINYGO) test -c -target=pybadge && rm smoke.test
	@go run ./tools/run-smoketest make smoketest-commands

smoketest-commands:
	# test all examples (except pwm)
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/adc
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/blinkm
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/blinky2
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/button
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/button2
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/echo
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/i2s
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/mcp3008
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/memstats
	$(TINYGO) build -size short -o test.hex -target=microbit            examples/microbit-blink
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/pininterrupt
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/serial
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/systick
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/test
	# test simulated boards on play.tinygo.org
ifneq ($(WASM), 0)
	$(TINYGO) build -size short -o test.wasm -tags=arduino              examples/blinky1
	$(TINYGO) build -size short -o test.wasm -tags=hifive1b             examples/blinky1
	$(TINYGO) build -size short -o test.wasm -tags=reelboard            examples/blinky1
	$(TINYGO) build -size short -o test.wasm -tags=pca10040             examples/blinky2
	$(TINYGO) build -size short -o test.wasm -tags=pca10056             examples/blinky2
	$(TINYGO) build -size short -o test.wasm -tags=circuitplay_express  examples/blinky1
endif
	# test all targets/boards
	$(TINYGO) build -size short -o test.hex -target=pca10040-s132v6     examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=microbit            examples/echo
	$(TINYGO) build -size short -o test.hex -target=microbit-s110v8     examples/echo
	$(TINYGO) build -size short -o test.hex -target=microbit-v2         examples/microbit-blink
	$(TINYGO) build -size short -o test.hex -target=microbit-v2-s113v7  examples/microbit-blink
	$(TINYGO) build -size short -o test.hex -target=nrf52840-mdk        examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=pca10031            examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=reelboard           examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=reelboard           examples/blinky2
	$(TINYGO) build -size short -o test.hex -target=pca10056            examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=pca10056            examples/blinky2
	$(TINYGO) build -size short -o test.hex -target=pca10059            examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=pca10059            examples/blinky2
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-m0        examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=feather-m0          examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=trinket-m0          examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=circuitplay-bluefruit examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/i2s
	$(TINYGO) build -size short -o test.hex -target=clue-alpha          examples/blinky1
	$(TINYGO) build -size short -o test.gba -target=gameboy-advance     examples/gba-display
	$(TINYGO) build -size short -o test.hex -target=grandcentral-m4     examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-m4        examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=feather-m4          examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=pybadge             examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=metro-m4-airlift    examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=pyportal            examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=particle-argon      examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=particle-boron      examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=particle-xenon      examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=pinetime-devkit0    examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=x9pro               examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=pca10056-s140v7     examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=reelboard-s140v7    examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=wioterminal         examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=pygamer             examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=xiao                examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/dac
	$(TINYGO) build -size short -o test.hex -target=pyportal            examples/dac
	$(TINYGO) build -size short -o test.hex -target=feather-nrf52840  	examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=feather-nrf52840-sense examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-nrf52840  examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=qtpy                examples/serial
	$(TINYGO) build -size short -o test.hex -target=teensy41            examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=teensy40            examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=teensy36            examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=p1am-100            examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=atsame54-xpro       examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=atsame54-xpro       examples/can
	$(TINYGO) build -size short -o test.hex -target=feather-m4-can      examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=feather-m4-can      examples/caninterrupt
	$(TINYGO) build -size short -o test.hex -target=arduino-nano33      examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=arduino-mkrwifi1010 examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=pico                examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=nano-33-ble         examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=nano-rp2040         examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=feather-rp2040 		examples/blinky1
	# test pwm
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-m0        examples/pwm
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-m4        examples/pwm
	$(TINYGO) build -size short -o test.hex -target=feather-m4          examples/pwm
ifneq ($(STM32), 0)
	$(TINYGO) build -size short -o test.hex -target=bluepill            examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=feather-stm32f405   examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=lgt92               examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=nucleo-f103rb       examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=nucleo-f722ze       examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=nucleo-l031k6       examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=nucleo-l432kc       examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=nucleo-l552ze       examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=nucleo-wl55jc       examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=stm32f4disco        examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=stm32f4disco        examples/blinky2
	$(TINYGO) build -size short -o test.hex -target=stm32f4disco-1      examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=stm32f4disco-1      examples/pwm
	$(TINYGO) build -size short -o test.hex -target=stm32f469disco      examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=lorae5              examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=swan                examples/blinky1
endif
ifneq ($(AVR), 0)
	$(TINYGO) build -size short -o test.hex -target=atmega1284p         examples/serial
	$(TINYGO) build -size short -o test.hex -target=arduino             examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=arduino             examples/pwm
	$(TINYGO) build -size short -o test.hex -target=arduino -scheduler=tasks  examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=arduino-mega1280    examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=arduino-mega1280    examples/pwm
	$(TINYGO) build -size short -o test.hex -target=arduino-nano        examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=digispark           examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=digispark -gc=leaking examples/blinky1
endif
ifneq ($(XTENSA), 0)
	$(TINYGO) build -size short -o test.bin -target=esp32-mini32      	examples/blinky1
	$(TINYGO) build -size short -o test.bin -target=nodemcu             examples/blinky1
	$(TINYGO) build -size short -o test.bin -target m5stack-core2       examples/serial
	$(TINYGO) build -size short -o test.bin -target m5stack             examples/serial
endif
	$(TINYGO) build -size short -o test.bin -target=esp32c3           	examples/serial
	$(TINYGO) build -size short -o test.bin -target=m5stamp-c3          examples/serial
	$(TINYGO) build -size short -o test.hex -target=hifive1b            examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=hifive1-qemu        examples/serial
	$(TINYGO) build -size short -o test.hex -target=maixbit             examples/blinky1
ifneq ($(WASM), 0)
	$(TINYGO) build -size short -o wasm.wasm -target=wasm               examples/wasm/export
	$(TINYGO) build -size short -o wasm.wasm -target=wasm               examples/wasm/main
endif
	# test various compiler flags
	$(TINYGO) build -size short -o test.hex -target=pca10040 -gc=none -scheduler=none examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=pca10040 -opt=1     examples/blinky1
	$(TINYGO) build -size short -o test.hex -target=pca10040 -serial=none examples/echo
	$(TINYGO) build             -o test.nro -target=nintendoswitch      examples/serial
	$(TINYGO) build -size short -o test.hex -target=pca10040 -opt=0     ./testdata/stdlib.go
	GOOS=linux GOARCH=arm $(TINYGO) build -size short -o test.elf       ./testdata/cgo
	GOOS=windows GOARCH=amd64 $(TINYGO) build -size short -o test.exe   ./testdata/cgo
	GOOS=darwin GOARCH=amd64 $(TINYGO) build              -o test       ./testdata/cgo
ifneq ($(OS),Windows_NT)
	# TODO: this does not yet work on Windows. Somehow, unused functions are
	# not garbage collected.
	$(TINYGO) build -o test.elf -gc=leaking -scheduler=none examples/serial
endif


wasmtest:
	$(GO) test ./tests/wasm

build/release: tinygo gen-device wasi-libc $(if $(filter 1,$(USE_SYSTEM_BINARYEN)),,binaryen)
	@mkdir -p build/release/tinygo/bin
	@mkdir -p build/release/tinygo/lib/clang/include
	@mkdir -p build/release/tinygo/lib/CMSIS/CMSIS
	@mkdir -p build/release/tinygo/lib/compiler-rt/lib
	@mkdir -p build/release/tinygo/lib/macos-minimal-sdk
	@mkdir -p build/release/tinygo/lib/mingw-w64/mingw-w64-crt/lib-common
	@mkdir -p build/release/tinygo/lib/mingw-w64/mingw-w64-headers/defaults
	@mkdir -p build/release/tinygo/lib/musl/arch
	@mkdir -p build/release/tinygo/lib/musl/crt
	@mkdir -p build/release/tinygo/lib/musl/src
	@mkdir -p build/release/tinygo/lib/nrfx
	@mkdir -p build/release/tinygo/lib/picolibc/newlib/libc
	@mkdir -p build/release/tinygo/lib/picolibc/newlib/libm
	@mkdir -p build/release/tinygo/lib/wasi-libc
	@mkdir -p build/release/tinygo/pkg/thumbv6m-unknown-unknown-eabi-cortex-m0
	@mkdir -p build/release/tinygo/pkg/thumbv6m-unknown-unknown-eabi-cortex-m0plus
	@mkdir -p build/release/tinygo/pkg/thumbv7em-unknown-unknown-eabi-cortex-m4
	@echo copying source files
	@cp -p  build/tinygo$(EXE)           build/release/tinygo/bin
ifneq ($(USE_SYSTEM_BINARYEN),1)
	@cp -p  build/wasm-opt$(EXE)         build/release/tinygo/bin
endif
	@cp -p $(abspath $(CLANG_SRC))/lib/Headers/*.h build/release/tinygo/lib/clang/include
	@cp -rp lib/CMSIS/CMSIS/Include      build/release/tinygo/lib/CMSIS/CMSIS
	@cp -rp lib/CMSIS/README.md          build/release/tinygo/lib/CMSIS
	@cp -rp lib/compiler-rt/lib/builtins build/release/tinygo/lib/compiler-rt/lib
	@cp -rp lib/compiler-rt/LICENSE.TXT  build/release/tinygo/lib/compiler-rt
	@cp -rp lib/compiler-rt/README.txt   build/release/tinygo/lib/compiler-rt
	@cp -rp lib/macos-minimal-sdk/*      build/release/tinygo/lib/macos-minimal-sdk
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
	./build/tinygo build-library -target=cortex-m0     -o build/release/tinygo/pkg/thumbv6m-unknown-unknown-eabi-cortex-m0/compiler-rt     compiler-rt
	./build/tinygo build-library -target=cortex-m0plus -o build/release/tinygo/pkg/thumbv6m-unknown-unknown-eabi-cortex-m0plus/compiler-rt compiler-rt
	./build/tinygo build-library -target=cortex-m4     -o build/release/tinygo/pkg/thumbv7em-unknown-unknown-eabi-cortex-m4/compiler-rt    compiler-rt
	./build/tinygo build-library -target=cortex-m0     -o build/release/tinygo/pkg/thumbv6m-unknown-unknown-eabi-cortex-m0/picolibc     picolibc
	./build/tinygo build-library -target=cortex-m0plus -o build/release/tinygo/pkg/thumbv6m-unknown-unknown-eabi-cortex-m0plus/picolibc picolibc
	./build/tinygo build-library -target=cortex-m4     -o build/release/tinygo/pkg/thumbv7em-unknown-unknown-eabi-cortex-m4/picolibc    picolibc

release: build/release
	tar -czf build/release.tar.gz -C build/release tinygo

deb: build/release
	@mkdir -p build/release-deb/usr/local/bin
	@mkdir -p build/release-deb/usr/local/lib
	cp -ar build/release/tinygo build/release-deb/usr/local/lib/tinygo
	ln -sf ../lib/tinygo/bin/tinygo build/release-deb/usr/local/bin/tinygo
	fpm -f -s dir -t deb -n tinygo -v $(shell grep "const Version = " goenv/version.go | awk '{print $$NF}') -m '@tinygo-org' --description='TinyGo is a Go compiler for small places.' --license='BSD 3-Clause' --url=https://tinygo.org/ --deb-changelog CHANGELOG.md -p build/release.deb -C ./build/release-deb
