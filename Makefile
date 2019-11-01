
# aliases
all: tinygo

# Default build and source directories, as created by `make llvm-build`.
LLVM_BUILDDIR ?= llvm-build
CLANG_SRC ?= llvm-project/clang
LLD_SRC ?= llvm-project/lld

# Go binary and GOROOT to select
GO ?= go
export GOROOT = $(shell $(GO) env GOROOT)

# md5sum binary
MD5SUM = md5sum

# Python binary
PYTHON ?= python

# tinygo binary for tests
TINYGO ?= tinygo

# Use CCACHE for LLVM if possible
ifneq (, $(shell which ccache))
    LLVM_OPTION += '-DLLVM_CCACHE_BUILD=ON'
endif

# Allow enabling LLVM assertions
ifeq (1, $(ASSERT))
    LLVM_OPTION += '-DLLVM_ENABLE_ASSERTIONS=ON'
else
    LLVM_OPTION += '-DLLVM_ENABLE_ASSERTIONS=OFF'
endif

.PHONY: all tinygo test $(LLVM_BUILDDIR) llvm-source clean fmt gen-device gen-device-nrf gen-device-avr

LLVM_COMPONENTS = all-targets analysis asmparser asmprinter bitreader bitwriter codegen core coroutines debuginfodwarf executionengine instrumentation interpreter ipo irreader linker lto mc mcjit objcarcopts option profiledata scalaropts support target

ifeq ($(OS),Windows_NT)
    EXE = .exe
    START_GROUP = -Wl,--start-group
    END_GROUP = -Wl,--end-group

    # LLVM compiled using MinGW on Windows appears to have problems with threads.
    # Without this flag, linking results in errors like these:
    #     libLLVMSupport.a(Threading.cpp.obj):Threading.cpp:(.text+0x55): undefined reference to `std::thread::hardware_concurrency()'
    LLVM_OPTION += -DLLVM_ENABLE_THREADS=OFF

    CGO_LDFLAGS += -static -static-libgcc -static-libstdc++
    CGO_LDFLAGS_EXTRA += -lversion

    # Build libclang manually because the CMake-based build system on Windows
    # doesn't allow building libclang as a static library.
    LIBCLANG_PATH = $(abspath build/libclang-custom.a)
    LIBCLANG_FILES = $(abspath $(wildcard $(LLVM_BUILDDIR)/tools/clang/tools/libclang/CMakeFiles/libclang.dir/*.cpp.obj))

    # Add the libclang dependency to the tinygo binary target.
tinygo: $(LIBCLANG_PATH)
test: $(LIBCLANG_PATH)
    # Build libclang.
$(LIBCLANG_PATH): $(LIBCLANG_FILES)
	@mkdir -p build
	ar rcs $(LIBCLANG_PATH) $^

else ifeq ($(shell uname -s),Darwin)
    MD5SUM = md5
    LIBCLANG_PATH = $(abspath $(LLVM_BUILDDIR))/lib/libclang.a
else
    LIBCLANG_PATH = $(abspath $(LLVM_BUILDDIR))/lib/libclang.a
    START_GROUP = -Wl,--start-group
    END_GROUP = -Wl,--end-group
endif

CLANG_LIBS = $(START_GROUP) -lclangAnalysis -lclangARCMigrate -lclangAST -lclangASTMatchers -lclangBasic -lclangCodeGen -lclangCrossTU -lclangDriver -lclangDynamicASTMatchers -lclangEdit -lclangFormat -lclangFrontend -lclangFrontendTool -lclangHandleCXX -lclangHandleLLVM -lclangIndex -lclangLex -lclangParse -lclangRewrite -lclangRewriteFrontend -lclangSema -lclangSerialization -lclangStaticAnalyzerCheckers -lclangStaticAnalyzerCore -lclangStaticAnalyzerFrontend -lclangTooling -lclangToolingASTDiff -lclangToolingCore -lclangToolingInclusions $(END_GROUP) -lstdc++

LLD_LIBS = $(START_GROUP) -llldCOFF -llldCommon -llldCore -llldDriver -llldELF -llldMachO -llldMinGW -llldReaderWriter -llldWasm -llldYAML $(END_GROUP)


# For static linking.
ifneq ("$(wildcard $(LLVM_BUILDDIR)/bin/llvm-config*)","")
    CGO_CPPFLAGS=$(shell $(LLVM_BUILDDIR)/bin/llvm-config --cppflags) -I$(abspath $(CLANG_SRC))/include -I$(abspath $(LLD_SRC))/include
    CGO_CXXFLAGS=-std=c++11
    CGO_LDFLAGS+=$(LIBCLANG_PATH) -std=c++11 -L$(abspath $(LLVM_BUILDDIR)/lib) $(CLANG_LIBS) $(LLD_LIBS) $(shell $(LLVM_BUILDDIR)/bin/llvm-config --ldflags --libs --system-libs $(LLVM_COMPONENTS)) -lstdc++ $(CGO_LDFLAGS_EXTRA)
endif


clean:
	@rm -rf build

FMT_PATHS = ./*.go builder cgo compiler interp ir loader src/device/arm src/examples src/machine src/os src/reflect src/runtime src/sync src/syscall src/internal/reflectlite transform
fmt:
	@gofmt -l -w $(FMT_PATHS)
fmt-check:
	@unformatted=$$(gofmt -l $(FMT_PATHS)); [ -z "$$unformatted" ] && exit 0; echo "Unformatted:"; for fn in $$unformatted; do echo "  $$fn"; done; exit 1


gen-device: gen-device-avr gen-device-nrf gen-device-sam gen-device-sifive gen-device-stm32

gen-device-avr:
	$(PYTHON) ./tools/gen-device-avr.py lib/avr/packs/atmega src/device/avr/
	$(PYTHON) ./tools/gen-device-avr.py lib/avr/packs/tiny src/device/avr/
	GO111MODULE=off $(GO) fmt ./src/device/avr

gen-device-nrf:
	$(PYTHON) ./tools/gen-device-svd.py lib/nrfx/mdk/ src/device/nrf/ --source=https://github.com/NordicSemiconductor/nrfx/tree/master/mdk
	GO111MODULE=off $(GO) fmt ./src/device/nrf

gen-device-sam:
	$(PYTHON) ./tools/gen-device-svd.py lib/cmsis-svd/data/Atmel/ src/device/sam/ --source=https://github.com/posborne/cmsis-svd/tree/master/data/Atmel
	GO111MODULE=off $(GO) fmt ./src/device/sam

gen-device-sifive:
	$(PYTHON) ./tools/gen-device-svd.py lib/cmsis-svd/data/SiFive-Community/ src/device/sifive/ --source=https://github.com/posborne/cmsis-svd/tree/master/data/SiFive-Community
	GO111MODULE=off $(GO) fmt ./src/device/sifive

gen-device-stm32:
	$(PYTHON) ./tools/gen-device-svd.py lib/cmsis-svd/data/STMicro/ src/device/stm32/ --source=https://github.com/posborne/cmsis-svd/tree/master/data/STMicro
	GO111MODULE=off $(GO) fmt ./src/device/stm32


# Get LLVM sources.
llvm-project/README.md:
	git clone -b release/9.x https://github.com/llvm/llvm-project
llvm-source: llvm-project/README.md

# Configure LLVM.
TINYGO_SOURCE_DIR=$(shell pwd)
$(LLVM_BUILDDIR)/build.ninja: llvm-source
	mkdir -p $(LLVM_BUILDDIR); cd $(LLVM_BUILDDIR); cmake -G Ninja $(TINYGO_SOURCE_DIR)/llvm-project/llvm "-DLLVM_TARGETS_TO_BUILD=X86;ARM;AArch64;RISCV;WebAssembly" "-DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD=AVR" -DCMAKE_BUILD_TYPE=Release -DLIBCLANG_BUILD_STATIC=ON -DLLVM_ENABLE_TERMINFO=OFF -DLLVM_ENABLE_ZLIB=OFF -DLLVM_ENABLE_PROJECTS="clang;lld" -DLLVM_TOOL_CLANG_TOOLS_EXTRA_BUILD=OFF $(LLVM_OPTION)

# Build LLVM.
$(LLVM_BUILDDIR): $(LLVM_BUILDDIR)/build.ninja
	cd $(LLVM_BUILDDIR); ninja


# Build the Go compiler.
tinygo:
	@if [ ! -f "$(LLVM_BUILDDIR)/bin/llvm-config" ]; then echo "Fetch and build LLVM first by running:"; echo "  make llvm-source"; echo "  make $(LLVM_BUILDDIR)"; exit 1; fi
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) build -o build/tinygo$(EXE) -tags byollvm .

test:
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test -v -tags byollvm ./cgo ./compileopts ./interp ./transform .

tinygo-test:
	cd tests/tinygotest && tinygo test

.PHONY: smoketest
smoketest:
	# test all examples
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
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/i2s
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/mcp3008
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=microbit            examples/microbit-blink
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/pwm
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/serial
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/systick
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10040            examples/test
	@$(MD5SUM) test.hex
	# test simulated boards on play.tinygo.org
	$(TINYGO) build             -o test.wasm -tags=arduino              examples/blinky1
	@$(MD5SUM) test.wasm
	$(TINYGO) build             -o test.wasm -tags=hifive1b             examples/blinky1
	@$(MD5SUM) test.wasm
	$(TINYGO) build             -o test.wasm -tags=reelboard            examples/blinky1
	@$(MD5SUM) test.wasm
	$(TINYGO) build             -o test.wasm -tags=pca10040             examples/blinky2
	@$(MD5SUM) test.wasm
	$(TINYGO) build             -o test.wasm -tags=pca10056             examples/blinky2
	@$(MD5SUM) test.wasm
	# test all targets/boards
	$(TINYGO) build -size short -o test.hex -target=pca10040-s132v6     examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=microbit            examples/echo
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nrf52840-mdk        examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10031            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=bluepill            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=reelboard           examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=reelboard           examples/blinky2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10056            examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pca10056            examples/blinky2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-m0        examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-m0          examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=trinket-m0          examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=stm32f4disco        examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=stm32f4disco        examples/blinky2
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=circuitplay-express examples/i2s
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.gba -target=gameboy-advance     examples/gba-display
	@$(MD5SUM) test.gba
	$(TINYGO) build -size short -o test.hex -target=itsybitsy-m4        examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=feather-m4          examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=nucleo-f103rb       examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=pinetime-devkit0    examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=x9pro               examples/blinky1
	@$(MD5SUM) test.hex
ifneq ($(AVR), 0)
	$(TINYGO) build -size short -o test.hex -target=arduino             examples/blinky1
	@$(MD5SUM) test.hex
	$(TINYGO) build -size short -o test.hex -target=digispark           examples/blinky1
	@$(MD5SUM) test.hex
endif
ifneq ($(RISCV), 0)
	$(TINYGO) build -size short -o test.hex -target=hifive1b            examples/blinky1
	@$(MD5SUM) test.hex
endif
	$(TINYGO) build             -o wasm.wasm -target=wasm               examples/wasm/export
	$(TINYGO) build             -o wasm.wasm -target=wasm               examples/wasm/main

release: tinygo gen-device
	@mkdir -p build/release/tinygo/bin
	@mkdir -p build/release/tinygo/lib/clang/include
	@mkdir -p build/release/tinygo/lib/CMSIS/CMSIS
	@mkdir -p build/release/tinygo/lib/compiler-rt/lib
	@mkdir -p build/release/tinygo/lib/nrfx
	@mkdir -p build/release/tinygo/pkg/armv6m-none-eabi
	@mkdir -p build/release/tinygo/pkg/armv7m-none-eabi
	@mkdir -p build/release/tinygo/pkg/armv7em-none-eabi
	@echo copying source files
	@cp -p  build/tinygo$(EXE)                     build/release/tinygo/bin
	@cp -p $(abspath $(CLANG_SRC))/lib/Headers/*.h build/release/tinygo/lib/clang/include
	@cp -rp lib/CMSIS/CMSIS/Include                build/release/tinygo/lib/CMSIS/CMSIS
	@cp -rp lib/CMSIS/README.md                    build/release/tinygo/lib/CMSIS
	@cp -rp llvm-project/compiler-rt/lib/builtins  build/release/tinygo/lib/compiler-rt/lib
	@cp -rp llvm-project/compiler-rt/LICENSE.TXT   build/release/tinygo/lib/compiler-rt
	@cp -rp llvm-project/compiler-rt/README.txt    build/release/tinygo/lib/compiler-rt
	@cp -rp lib/nrfx/*                             build/release/tinygo/lib/nrfx
	@cp -rp src                                    build/release/tinygo/src
	@cp -rp targets                                build/release/tinygo/targets
	./build/tinygo build-builtins -target=armv6m-none-eabi  -o build/release/tinygo/pkg/armv6m-none-eabi/compiler-rt.a
	./build/tinygo build-builtins -target=armv7m-none-eabi  -o build/release/tinygo/pkg/armv7m-none-eabi/compiler-rt.a
	./build/tinygo build-builtins -target=armv7em-none-eabi -o build/release/tinygo/pkg/armv7em-none-eabi/compiler-rt.a
	tar -czf build/release.tar.gz -C build/release tinygo
