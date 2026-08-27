# Build configuration: host detection, LLVM paths, and CGO flags.


# Default build and source directories, as created by `make llvm-build`.
LLVM_BUILDDIR ?= llvm-build
LLVM_PROJECTDIR ?= llvm-project
CLANG_SRC ?= $(LLVM_PROJECTDIR)/clang
LLD_SRC ?= $(LLVM_PROJECTDIR)/lld

ifeq ($(OS),Windows_NT)
    # avoid calling uname on Windows
    uname := Windows_NT
else
    uname := $(shell uname -s)
endif

# Try to autodetect LLVM build tools.
# Versions are listed here in descending priority order.
LLVM_VERSIONS = 19 18 17 16 15
errifempty = $(if $(1),$(1),$(error $(2)))
detect = $(shell which $(call errifempty,$(firstword $(foreach p,$(2),$(shell command -v $(p) 2> /dev/null && echo $(p)))),failed to locate $(1) at any of: $(2)))
toolSearchPathsVersion = $(1)-$(2)
ifeq ($(uname),Darwin)
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

# Flags to pass to go test.
GOTESTFLAGS ?=
GOTESTPKGS ?= ./builder ./cgo ./compileopts ./compiler ./interp ./transform .

# tinygo binary for tests
TINYGO ?= $(call detect,tinygo,tinygo $(CURDIR)/build/tinygo)

# Check for ccache if the user hasn't set it to on or off.
ifeq (, $(CCACHE))
    LLVM_OPTION += '-DLLVM_CCACHE_BUILD=$(if $(shell command -v ccache 2> /dev/null),ON,OFF)'
else
    LLVM_OPTION += '-DLLVM_CCACHE_BUILD=$(CCACHE)'
endif

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

# Optimize the binary size for Linux.
# These flags may work on other platforms, but have only been tested on Linux.
ifeq ($(uname),Linux)
    HAS_MOLD := $(shell command -v ld.mold 2> /dev/null)
    HAS_LLD := $(shell command -v ld.lld 2> /dev/null)
    LLVM_CFLAGS := -ffunction-sections -fdata-sections -fvisibility=hidden
    LLVM_LDFLAGS := -Wl,--gc-sections
    ifneq ($(HAS_MOLD),)
        # Mold might be slightly faster.
        LLVM_LDFLAGS += -fuse-ld=mold -Wl,--icf=all
    else ifneq ($(HAS_LLD),)
        # LLD is more commonly available.
        LLVM_LDFLAGS += -fuse-ld=lld -Wl,--icf=all
    endif
    LLVM_OPTION += \
        -DCMAKE_C_FLAGS="$(LLVM_CFLAGS)" \
        -DCMAKE_CXX_FLAGS="$(LLVM_CFLAGS)"
    CGO_LDFLAGS += $(LLVM_LDFLAGS)
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

