
# aliases
all: tinygo

# Default build and source directories, as created by `make llvm-build`.
LLVM_BUILDDIR ?= llvm-build
LLVM_PROJECTDIR ?= llvm-project
CLANG_SRC ?= $(LLVM_PROJECTDIR)/clang
LLD_SRC ?= $(LLVM_PROJECTDIR)/lld

# Try to autodetect LLVM build tools.
# Versions are listed here in descending priority order.
LLVM_VERSIONS = 17 16 15
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
GOTESTFLAGS ?=

# tinygo binary for tests
TINYGO ?= $(call detect,tinygo,tinygo $(CURDIR)/build/tinygo)

# Check for ccache if the user hasn't set it to on or off.
ifeq (, $(CCACHE))
    # Use CCACHE for LLVM if possible
    ifneq (, $(shell command -v ccache 2> /dev/null))
        CCACHE := ON
    else
        CCACHE := OFF
    endif
endif
LLVM_OPTION += '-DLLVM_CCACHE_BUILD=$(CCACHE)'

# Allow enabling LLVM assertions
ifeq (1, $(ASSERT))
    LLVM_OPTION += '-DLLVM_ENABLE_ASSERTIONS=ON'
else
    LLVM_OPTION += '-DLLVM_ENABLE_ASSERTIONS=OFF'
endif

# Enable AddressSanitizer
ifeq (1, $(ASAN))
    LLVM_OPTION += -DLLVM_USE_SANITIZER=Address
    CGO_LDFLAGS += -fsanitize=address
endif

ifeq (1, $(STATIC))
    # Build TinyGo as a fully statically linked binary (no dynamically loaded
    # libraries such as a libc). This is not supported with glibc which is used
    # on most major Linux distributions. However, it is supported in Alpine
    # Linux with musl.
    CGO_LDFLAGS += -static
    # Also set the thread stack size to 1MB. This is necessary on musl as the
    # default stack size is 128kB and LLVM uses more than that.
    # For more information, see:
    # https://wiki.musl-libc.org/functional-differences-from-glibc.html#Thread-stack-size
    CGO_LDFLAGS += -Wl,-z,stack-size=1048576
    # Build wasm-opt with static linking.
    # For details, see:
    # https://github.com/WebAssembly/binaryen/blob/version_102/.github/workflows/ci.yml#L181
    BINARYEN_OPTION += -DCMAKE_CXX_FLAGS="-static" -DCMAKE_C_FLAGS="-static"
endif

# Cross compiling support.
ifneq ($(CROSS),)
    CC = $(CROSS)-gcc
    CXX = $(CROSS)-g++
    LLVM_OPTION += \
        -DCMAKE_C_COMPILER=$(CC) \
        -DCMAKE_CXX_COMPILER=$(CXX) \
        -DLLVM_DEFAULT_TARGET_TRIPLE=$(CROSS) \
        -DCROSS_TOOLCHAIN_FLAGS_NATIVE="-UCMAKE_C_COMPILER;-UCMAKE_CXX_COMPILER"
    ifeq ($(CROSS), arm-linux-gnueabihf)
        # Assume we're building on a Debian-like distro, with QEMU installed.
        LLVM_CONFIG_PREFIX = qemu-arm -L /usr/arm-linux-gnueabihf/
        # The CMAKE_SYSTEM_NAME flag triggers cross compilation mode.
        LLVM_OPTION += \
            -DCMAKE_SYSTEM_NAME=Linux \
            -DLLVM_TARGET_ARCH=ARM
        GOENVFLAGS = GOARCH=arm CC=$(CC) CXX=$(CXX) CGO_ENABLED=1
        BINARYEN_OPTION += -DCMAKE_C_COMPILER=$(CC) -DCMAKE_CXX_COMPILER=$(CXX)
    else ifeq ($(CROSS), aarch64-linux-gnu)
        # Assume we're building on a Debian-like distro, with QEMU installed.
        LLVM_CONFIG_PREFIX = qemu-aarch64 -L /usr/aarch64-linux-gnu/
        # The CMAKE_SYSTEM_NAME flag triggers cross compilation mode.
        LLVM_OPTION += \
            -DCMAKE_SYSTEM_NAME=Linux \
            -DLLVM_TARGET_ARCH=AArch64
        GOENVFLAGS = GOARCH=arm64 CC=$(CC) CXX=$(CXX) CGO_ENABLED=1
        BINARYEN_OPTION += -DCMAKE_C_COMPILER=$(CC) -DCMAKE_CXX_COMPILER=$(CXX)
    else
        $(error Unknown cross compilation target: $(CROSS))
    endif
endif

.PHONY: all tinygo test $(LLVM_BUILDDIR) llvm-source clean fmt gen-device gen-device-nrf gen-device-nxp gen-device-avr gen-device-rp

LLVM_COMPONENTS = all-targets analysis asmparser asmprinter bitreader bitwriter codegen core coroutines coverage debuginfodwarf debuginfopdb executionengine frontendhlsl frontendopenmp instrumentation interpreter ipo irreader libdriver linker lto mc mcjit objcarcopts option profiledata scalaropts support target windowsdriver windowsmanifest

ifeq ($(OS),Windows_NT)
    EXE = .exe
    START_GROUP = -Wl,--start-group
    END_GROUP = -Wl,--end-group

    # PIC needs to be disabled for libclang to work.
    LLVM_OPTION += -DLLVM_ENABLE_PIC=OFF

    CGO_CPPFLAGS += -DCINDEX_NO_EXPORTS
    CGO_LDFLAGS += -static -static-libgcc -static-libstdc++
    CGO_LDFLAGS_EXTRA += -lversion

    USE_SYSTEM_BINARYEN ?= 1

else ifeq ($(shell uname -s),Darwin)
    MD5SUM ?= md5

    CGO_LDFLAGS += -lxar

    USE_SYSTEM_BINARYEN ?= 1

else ifeq ($(shell uname -s),FreeBSD)
    MD5SUM ?= md5
    START_GROUP = -Wl,--start-group
    END_GROUP = -Wl,--end-group
else
    START_GROUP = -Wl,--start-group
    END_GROUP = -Wl,--end-group
endif

# md5sum binary default, can be overridden by an environment variable
MD5SUM ?= md5sum

# Libraries that should be linked in for the statically linked Clang.
CLANG_LIB_NAMES = clangAnalysis clangAST clangASTMatchers clangBasic clangCodeGen clangCrossTU clangDriver clangDynamicASTMatchers clangEdit clangExtractAPI clangFormat clangFrontend clangFrontendTool clangHandleCXX clangHandleLLVM clangIndex clangLex clangParse clangRewrite clangRewriteFrontend clangSema clangSerialization clangSupport clangTooling clangToolingASTDiff clangToolingCore clangToolingInclusions
CLANG_LIBS = $(START_GROUP) $(addprefix -l,$(CLANG_LIB_NAMES)) $(END_GROUP) -lstdc++

# Libraries that should be linked in for the statically linked LLD.
LLD_LIB_NAMES = lldCOFF lldCommon lldELF lldMachO lldMinGW lldWasm
LLD_LIBS = $(START_GROUP) $(addprefix -l,$(LLD_LIB_NAMES)) $(END_GROUP)

# Other libraries that are needed to link TinyGo.
EXTRA_LIB_NAMES = LLVMInterpreter LLVMMCA LLVMRISCVTargetMCA LLVMX86TargetMCA

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
    CGO_CPPFLAGS+=$(shell $(LLVM_CONFIG_PREFIX) $(LLVM_BUILDDIR)/bin/llvm-config --cppflags) -I$(abspath $(LLVM_BUILDDIR))/tools/clang/include -I$(abspath $(CLANG_SRC))/include -I$(abspath $(LLD_SRC))/include
    CGO_CXXFLAGS=-std=c++17
    CGO_LDFLAGS+=-L$(abspath $(LLVM_BUILDDIR)/lib) -lclang $(CLANG_LIBS) $(LLD_LIBS) $(shell $(LLVM_CONFIG_PREFIX) $(LLVM_BUILDDIR)/bin/llvm-config --ldflags --libs --system-libs $(LLVM_COMPONENTS)) -lstdc++ $(CGO_LDFLAGS_EXTRA)
endif

clean:
	@rm -rf build

FMT_PATHS = ./*.go builder cgo/*.go compiler interp loader src transform
fmt:
	@gofmt -l -w $(FMT_PATHS)
fmt-check:
	@unformatted=$$(gofmt -l $(FMT_PATHS)); [ -z "$$unformatted" ] && exit 0; echo "Unformatted:"; for fn in $$unformatted; do echo "  $$fn"; done; exit 1


gen-device: gen-device-avr gen-device-esp gen-device-nrf gen-device-sam gen-device-sifive gen-device-kendryte gen-device-nxp gen-device-rp
ifneq ($(STM32), 0)
gen-device: gen-device-stm32
endif

gen-device-avr:
	@if [ ! -e lib/avr/README.md ]; then echo "Submodules have not been downloaded. Please download them using:\n  git submodule update --init --recursive"; exit 1; fi
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

gen-device-renesas: build/gen-device-svd
	./build/gen-device-svd -source=https://github.com/tinygo-org/renesas-svd lib/renesas-svd/ src/device/renesas/
	GO111MODULE=off $(GO) fmt ./src/device/renesas

# Get LLVM sources.
$(LLVM_PROJECTDIR)/llvm:
	git clone -b xtensa_release_16.x --depth=1 https://github.com/espressif/llvm-project $(LLVM_PROJECTDIR)
llvm-source: $(LLVM_PROJECTDIR)/llvm

# Configure LLVM.
TINYGO_SOURCE_DIR=$(shell pwd)
$(LLVM_BUILDDIR)/build.ninja:
	mkdir -p $(LLVM_BUILDDIR) && cd $(LLVM_BUILDDIR) && cmake -G Ninja $(TINYGO_SOURCE_DIR)/$(LLVM_PROJECTDIR)/llvm "-DLLVM_TARGETS_TO_BUILD=X86;ARM;AArch64;RISCV;WebAssembly" "-DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD=AVR;Xtensa" -DCMAKE_BUILD_TYPE=Release -DLIBCLANG_BUILD_STATIC=ON -DLLVM_ENABLE_TERMINFO=OFF -DLLVM_ENABLE_ZLIB=OFF -DLLVM_ENABLE_ZSTD=OFF -DLLVM_ENABLE_LIBEDIT=OFF -DLLVM_ENABLE_Z3_SOLVER=OFF -DLLVM_ENABLE_OCAMLDOC=OFF -DLLVM_ENABLE_LIBXML2=OFF -DLLVM_ENABLE_PROJECTS="clang;lld" -DLLVM_TOOL_CLANG_TOOLS_EXTRA_BUILD=OFF -DCLANG_ENABLE_STATIC_ANALYZER=OFF -DCLANG_ENABLE_ARCMT=OFF $(LLVM_OPTION)

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
	@if [ ! -e lib/wasi-libc/Makefile ]; then echo "Submodules have not been downloaded. Please download them using:\n  git submodule update --init --recursive"; exit 1; fi
	cd lib/wasi-libc && $(MAKE) -j4 EXTRA_CFLAGS="-O2 -g -DNDEBUG -mnontrapping-fptoint -msign-ext" MALLOC_IMPL=none CC="$(CLANG)" AR=$(LLVM_AR) NM=$(LLVM_NM)

# Check for Node.js used during WASM tests.
NODEJS_VERSION := $(word 1,$(subst ., ,$(shell node -v | cut -c 2-)))
MIN_NODEJS_VERSION=16

.PHONY: check-nodejs-version
check-nodejs-version:
ifeq (, $(shell which node))
	@echo "Install NodeJS version 16+ to run tests."; exit 1;
endif
	@if [ $(NODEJS_VERSION) -lt $(MIN_NODEJS_VERSION) ]; then echo "Install NodeJS version 16+ to run tests."; exit 1; fi

# Build the Go compiler.
tinygo:
	@if [ ! -f "$(LLVM_BUILDDIR)/bin/llvm-config" ]; then echo "Fetch and build LLVM first by running:"; echo "  $(MAKE) llvm-source"; echo "  $(MAKE) $(LLVM_BUILDDIR)"; exit 1; fi
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GOENVFLAGS) $(GO) build -buildmode exe -o build/tinygo$(EXE) -tags "byollvm osusergo" -ldflags="-X github.com/tinygo-org/tinygo/goenv.GitSha1=`git rev-parse --short HEAD`" .
test: wasi-libc check-nodejs-version
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=20m -buildmode exe -tags "byollvm osusergo" ./builder ./cgo ./compileopts ./compiler ./interp ./transform .

# Standard library packages that pass tests on darwin, linux, wasi, and windows, but take over a minute in wasi
TEST_PACKAGES_SLOW = \
	compress/bzip2 \
	crypto/dsa \
	index/suffixarray \

# Standard library packages that pass tests quickly on darwin, linux, wasi, and windows
TEST_PACKAGES_FAST = \
	compress/lzw \
	compress/zlib \
	container/heap \
	container/list \
	container/ring \
	crypto/des \
	crypto/md5 \
	crypto/rc4 \
	crypto/sha1 \
	crypto/sha256 \
	crypto/sha512 \
	debug/macho \
	embed/internal/embedtest \
	encoding \
	encoding/ascii85 \
	encoding/base32 \
	encoding/base64 \
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
	$(nil)

# Assume this will go away before Go2, so only check minor version.
ifeq ($(filter $(shell $(GO) env GOVERSION | cut -f 2 -d.), 16 17 18), )
TEST_PACKAGES_FAST += crypto/internal/nistec/fiat
else
TEST_PACKAGES_FAST += crypto/elliptic/internal/fiat
endif

# archive/zip requires os.ReadAt, which is not yet supported on windows
# bytes requires mmap
# compress/flate appears to hang on wasi
# crypto/hmac fails on wasi, it exits with a "slice out of range" panic
# debug/plan9obj requires os.ReadAt, which is not yet supported on windows
# image requires recover(), which is not  yet supported on wasi
# io/ioutil requires os.ReadDir, which is not yet supported on windows or wasi
# mime/quotedprintable requires syscall.Faccessat
# strconv requires recover() which is not yet supported on wasi
# text/tabwriter requries recover(), which is not  yet supported on wasi
# text/template/parse requires recover(), which is not yet supported on wasi
# testing/fstest requires os.ReadDir, which is not yet supported on windows or wasi

# Additional standard library packages that pass tests on individual platforms
TEST_PACKAGES_LINUX := \
	archive/zip \
	bytes \
	compress/flate \
	crypto/hmac \
	debug/dwarf \
	debug/plan9obj \
	image \
	io/ioutil \
	mime/quotedprintable \
	net \
	strconv \
	testing/fstest \
	text/tabwriter \
	text/template/parse

TEST_PACKAGES_DARWIN := $(TEST_PACKAGES_LINUX)

TEST_PACKAGES_WINDOWS := \
	compress/flate \
	crypto/hmac \
	strconv \
	text/template/parse \
	$(nil)

# Report platforms on which each standard library package is known to pass tests
jointmp := $(shell echo /tmp/join.$$$$)
report-stdlib-tests-pass:
	@for t in $(TEST_PACKAGES_DARWIN); do echo "$$t darwin"; done | sort > $(jointmp).darwin
	@for t in $(TEST_PACKAGES_LINUX); do echo "$$t linux"; done | sort > $(jointmp).linux
	@for t in $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_SLOW); do echo "$$t darwin linux wasi windows"; done | sort > $(jointmp).portable
	@join -a1 -a2 $(jointmp).darwin $(jointmp).linux | \
	join -a1 -a2 - $(jointmp).portable
	@rm $(jointmp).*

# Standard library packages that pass tests quickly on the current platform
ifeq ($(shell uname),Darwin)
TEST_PACKAGES_HOST := $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_DARWIN)
TEST_IOFS := true
endif
ifeq ($(shell uname),Linux)
TEST_PACKAGES_HOST := $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_LINUX)
TEST_IOFS := true
endif
ifeq ($(OS),Windows_NT)
TEST_PACKAGES_HOST := $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_WINDOWS)
TEST_IOFS := false
endif

# Test known-working standard library packages.
# TODO: parallelize, and only show failing tests (no implied -v flag).
.PHONY: tinygo-test
tinygo-test:
	$(TINYGO) test $(TEST_PACKAGES_HOST) $(TEST_PACKAGES_SLOW)
	@# io/fs requires os.ReadDir, not yet supported on windows or wasi. It also
	@# requires a large stack-size. Hence, io/fs is only run conditionally.
	@# For more details, see the comments on issue #3143.
ifeq ($(TEST_IOFS),true)
	$(TINYGO) test -stack-size=6MB io/fs
endif
tinygo-test-fast:
	$(TINYGO) test $(TEST_PACKAGES_HOST)
tinygo-bench:
	$(TINYGO) test -bench . $(TEST_PACKAGES_HOST) $(TEST_PACKAGES_SLOW)
tinygo-bench-fast:
	$(TINYGO) test -bench . $(TEST_PACKAGES_HOST)

# Same thing, except for wasi rather than the current platform.
tinygo-test-wasi:
	$(TINYGO) test -target wasi $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_SLOW) ./tests/runtime_wasi
tinygo-test-wasip1:
	GOOS=wasip1 GOARCH=wasm $(TINYGO) test $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_SLOW) ./tests/runtime_wasi
tinygo-test-wasi-fast:
	$(TINYGO) test -target wasi $(TEST_PACKAGES_FAST) ./tests/runtime_wasi
tinygo-test-wasip1-fast:
	GOOS=wasip1 GOARCH=wasm $(TINYGO) test $(TEST_PACKAGES_FAST) ./tests/runtime_wasi
tinygo-bench-wasi:
	$(TINYGO) test -target wasi -bench . $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_SLOW)
tinygo-bench-wasi-fast:
	$(TINYGO) test -target wasi -bench . $(TEST_PACKAGES_FAST)

# Test external packages in a large corpus.
test-corpus:
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=1h -buildmode exe -tags byollvm -run TestCorpus . -corpus=testdata/corpus.yaml
test-corpus-fast:
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=1h -buildmode exe -tags byollvm -run TestCorpus -short . -corpus=testdata/corpus.yaml
test-corpus-wasi: wasi-libc
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=1h -buildmode exe -tags byollvm -run TestCorpus . -corpus=testdata/corpus.yaml -target=wasi

tinygo-baremetal:
	# Regression tests that run on a baremetal target and don't fit in either main_test.go or smoketest.
	# regression test for #2666: e.g. encoding/hex must pass on baremetal
	$(TINYGO) test -target cortex-m-qemu encoding/hex

.PHONY: smoketest
smoketest:
	$(TINYGO) version
	$(TINYGO) targets > /dev/null
	# regression test for #2892
	cd tests/testing/recurse && ($(TINYGO) test ./... > recurse.log && cat recurse.log && test $$(wc -l < recurse.log) = 2 && rm recurse.log)
	# compile-only platform-independent examples
	cd tests/text/template/smoke && $(TINYGO) test -c && rm -f smoke.test
	# regression test for #2563
	cd tests/os/smoke && $(TINYGO) test -c -target=pybadge && rm smoke.test
	# test all examples (except pwm)
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/adc
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/blinkm
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/blinky2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/button
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/button2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/echo2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/i2s
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/mcp3008
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/memstats
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=microbit            examples/microbit-blink
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/pininterrupt
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nano-rp2040         examples/rtcinterrupt
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/serial
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/systick
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/test
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/time-offset
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=wioterminal         examples/hid-mouse
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=wioterminal         examples/hid-keyboard
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-rp2040      examples/i2c-target
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-rp2040      examples/watchdog
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-rp2040      examples/device-id
	@$(MD5SUM) test.hex
	# test simulated boards on play.tinygo.org
ifneq ($(WASM), 0)
	$(TINYGO) build -size short -o test.wasm -tags=arduino              examples/blinky1
	@$(MD5SUM) test.wasm
	$(TINYGO) build -size short -o test.wasm -tags=hifive1b             examples/blinky1
	@$(MD5SUM) test.wasm
	$(TINYGO) build -size short -o test.wasm -tags=reelboard            examples/blinky1
	@$(MD5SUM) test.wasm
	$(TINYGO) build -size short -o test.wasm -tags=microbit             examples/microbit-blink
	@$(MD5SUM) test.wasm
	$(TINYGO) build -size short -o test.wasm -tags=circuitplay_express  examples/blinky1
	@$(MD5SUM) test.wasm
	$(TINYGO) build -size short -o test.wasm -tags=circuitplay_bluefruit examples/blinky1
	@$(MD5SUM) test.wasm
	$(TINYGO) build -size short -o test.wasm -tags=mch2022              examples/serial
	@$(MD5SUM) test.wasm
endif
	# test all targets/boards
	$(TINYGO) build -size short -o test.hex -target=pca10040-s132v6     examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=microbit            examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=microbit-s110v8     examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=microbit-v2         examples/microbit-blink
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=microbit-v2-s113v7  examples/microbit-blink
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nrf52840-mdk        examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10031            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=reelboard           examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=reelboard           examples/blinky2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10056            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10056            examples/blinky2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10059            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10059            examples/blinky2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=bluemicro840        examples/blinky2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-m0        examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-m0          examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=trinket-m0          examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=gemma-m0            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=circuitplay-bluefruit examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/i2s
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=clue-alpha          examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.gba -target=gameboy-advance     examples/gba-display
	@$(MD5SUM) test.gba
	$(TINYGO) build -size short -o test.hex -target=grandcentral-m4     examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-m4        examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-m4          examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=matrixportal-m4     examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pybadge             examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=metro-m4-airlift    examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pyportal            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=particle-argon      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=particle-boron      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=particle-xenon      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pinetime            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=x9pro               examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10056-s140v7     examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=reelboard-s140v7    examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=wioterminal         examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pygamer             examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=xiao                examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/dac
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pyportal            examples/dac
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-nrf52840  	examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-nrf52840-sense examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-nrf52840  examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=qtpy                examples/serial
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=teensy41            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=teensy40            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=teensy36            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=p1am-100            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=atsame54-xpro       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=atsame54-xpro       examples/can
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-m4-can      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-m4-can      examples/caninterrupt
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-nano33      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-mkrwifi1010 examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pico                examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nano-33-ble         examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nano-rp2040         examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-rp2040 		examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=qtpy-rp2040         examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=kb2040              examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=macropad-rp2040 	examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=badger2040          examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=tufty2040           examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=thingplus-rp2040    examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=xiao-rp2040         examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=waveshare-rp2040-zero examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=challenger-rp2040    examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=trinkey-qt2040      examples/temp
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=gopher-badge      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=ae-rp2040           examples/echo
	@$(MD5SUM) test.hex
	# test pwm
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-m0        examples/pwm
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-m4        examples/pwm
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-m4          examples/pwm
	@$(MD5SUM) test.hex
	# test usb
	$(TINYGO) build -size short -o test.hex -target=feather-nrf52840    examples/hid-keyboard
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/hid-keyboard
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-nrf52840    examples/usb-midi
	@$(MD5SUM) test.hex
ifneq ($(STM32), 0)
	$(TINYGO) build -size short -o test.hex -target=bluepill            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-stm32f405   examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=lgt92               examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nucleo-f103rb       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nucleo-f722ze       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nucleo-l031k6       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nucleo-l432kc       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nucleo-l552ze       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nucleo-wl55jc       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=stm32f4disco        examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=stm32f4disco        examples/blinky2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=stm32f4disco-1      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=stm32f4disco-1      examples/pwm
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=stm32f469disco      examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=lorae5              examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=swan                examples/blinky1
	@$(MD5SUM) test.hex
endif
	$(TINYGO) build -size short -o test.hex -target=atmega1284p         examples/serial
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino             examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-leonardo    examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino             examples/pwm
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino -scheduler=tasks  examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-mega1280    examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-mega1280    examples/pwm
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=arduino-nano        examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=attiny1616          examples/empty
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=digispark           examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=digispark -gc=leaking examples/blinky1
	@$(MD5SUM) test.hex
ifneq ($(XTENSA), 0)
	$(TINYGO) build -size short -o test.bin -target=esp32-mini32      	examples/blinky1
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=nodemcu             examples/blinky1
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target m5stack-core2       examples/serial
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target m5stack             examples/serial
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target m5stick-c           examples/serial
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target mch2022             examples/serial
	@$(MD5SUM) test.bin
endif
	$(TINYGO) build -size short -o test.bin -target=esp32c3             examples/serial
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=esp32c3-12f         examples/serial
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=m5stamp-c3          examples/serial
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.bin -target=xiao-esp32c3        examples/serial
	@$(MD5SUM) test.bin
	$(TINYGO) build -size short -o test.hex -target=hifive1b            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=maixbit             examples/blinky1
	@$(MD5SUM) test.hex
ifneq ($(WASM), 0)
	$(TINYGO) build -size short -o wasm.wasm -target=wasm               examples/wasm/export
	$(TINYGO) build -size short -o wasm.wasm -target=wasm               examples/wasm/main
endif
	$(TINYGO) build -size short -o test.exe -target=uefi-amd64          examples/empty
	@$(MD5SUM) test.exe
	$(TINYGO) build -size short -o test.exe -target=uefi-amd64          examples/time-offset
	@$(MD5SUM) test.exe
	# test various compiler flags
	$(TINYGO) build -size short -o test.hex -target=pca10040 -gc=none -scheduler=none examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040 -opt=1     examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040 -serial=none examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build             -o test.nro -target=nintendoswitch      examples/serial
	@$(MD5SUM) test.nro
	$(TINYGO) build -size short -o test.hex -target=pca10040 -opt=0     ./testdata/stdlib.go
	@$(MD5SUM) test.hex
	GOOS=linux GOARCH=arm $(TINYGO) build -size short -o test.elf       ./testdata/cgo
	GOOS=windows GOARCH=amd64 $(TINYGO) build -size short -o test.exe   ./testdata/cgo
	GOOS=windows GOARCH=arm64 $(TINYGO) build -size short -o test.exe   ./testdata/cgo
	GOOS=darwin GOARCH=amd64 $(TINYGO) build  -size short -o test       ./testdata/cgo
	GOOS=darwin GOARCH=arm64 $(TINYGO) build  -size short -o test       ./testdata/cgo
	$(TINYGO) build -size short -o test.exe -target=uefi-amd64          ./testdata/cgo
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
	@cp -rp lib/musl/src/legacy          build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/malloc          build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/mman            build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/math            build/release/tinygo/lib/musl/src
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
	@cp -rp lib/picolibc/newlib/libm/math        build/release/tinygo/lib/picolibc/newlib/libm
	@cp -rp lib/picolibc-stdio.c         build/release/tinygo/lib
	@cp -rp lib/wasi-libc/sysroot        build/release/tinygo/lib/wasi-libc/sysroot
	@cp -rp llvm-project/compiler-rt/lib/builtins build/release/tinygo/lib/compiler-rt-builtins
	@cp -rp llvm-project/compiler-rt/LICENSE.TXT  build/release/tinygo/lib/compiler-rt-builtins
	@cp -rp src                          build/release/tinygo/src
	@cp -rp targets                      build/release/tinygo/targets
	./build/release/tinygo/bin/tinygo build-library -target=cortex-m0     -o build/release/tinygo/pkg/thumbv6m-unknown-unknown-eabi-cortex-m0/compiler-rt     compiler-rt
	./build/release/tinygo/bin/tinygo build-library -target=cortex-m0plus -o build/release/tinygo/pkg/thumbv6m-unknown-unknown-eabi-cortex-m0plus/compiler-rt compiler-rt
	./build/release/tinygo/bin/tinygo build-library -target=cortex-m4     -o build/release/tinygo/pkg/thumbv7em-unknown-unknown-eabi-cortex-m4/compiler-rt    compiler-rt
	./build/release/tinygo/bin/tinygo build-library -target=cortex-m0     -o build/release/tinygo/pkg/thumbv6m-unknown-unknown-eabi-cortex-m0/picolibc     picolibc
	./build/release/tinygo/bin/tinygo build-library -target=cortex-m0plus -o build/release/tinygo/pkg/thumbv6m-unknown-unknown-eabi-cortex-m0plus/picolibc picolibc
	./build/release/tinygo/bin/tinygo build-library -target=cortex-m4     -o build/release/tinygo/pkg/thumbv7em-unknown-unknown-eabi-cortex-m4/picolibc    picolibc

release:
	tar -czf build/release.tar.gz -C build/release tinygo

DEB_ARCH ?= native
deb:
	@mkdir -p build/release-deb/usr/local/bin
	@mkdir -p build/release-deb/usr/local/lib
	cp -ar build/release/tinygo build/release-deb/usr/local/lib/tinygo
	ln -sf ../lib/tinygo/bin/tinygo build/release-deb/usr/local/bin/tinygo
	fpm -f -s dir -t deb -n tinygo -a $(DEB_ARCH) -v $(shell grep "const version = " goenv/version.go | awk '{print $$NF}') -m '@tinygo-org' --description='TinyGo is a Go compiler for small places.' --license='BSD 3-Clause' --url=https://tinygo.org/ --deb-changelog CHANGELOG.md -p build/release.deb -C ./build/release-deb

ifneq ($(RELEASEONLY), 1)
release: build/release
deb: build/release
endif
