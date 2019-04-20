
# aliases
all: tinygo
tinygo: build/tinygo

.PHONY: all tinygo build/tinygo test llvm-build llvm-source clean fmt gen-device gen-device-nrf gen-device-avr

# Default build and source directories, as created by `make llvm-build`.
LLVM_BUILDDIR ?= llvm-build
CLANG_SRC ?= llvm/tools/clang
LLD_SRC ?= llvm/tools/lld

LLVM_COMPONENTS = all-targets analysis asmparser asmprinter bitreader bitwriter codegen core coroutines debuginfodwarf executionengine instrumentation interpreter ipo irreader linker lto mc mcjit objcarcopts option profiledata scalaropts support target

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    START_GROUP = -Wl,--start-group
    END_GROUP = -Wl,--end-group
endif

CLANG_LIBS = $(START_GROUP) $(abspath $(LLVM_BUILDDIR))/lib/libclang.a -lclangAnalysis -lclangARCMigrate -lclangAST -lclangASTMatchers -lclangBasic -lclangCodeGen -lclangCrossTU -lclangDriver -lclangDynamicASTMatchers -lclangEdit -lclangFormat -lclangFrontend -lclangFrontendTool -lclangHandleCXX -lclangHandleLLVM -lclangIndex -lclangLex -lclangParse -lclangRewrite -lclangRewriteFrontend -lclangSema -lclangSerialization -lclangStaticAnalyzerCheckers -lclangStaticAnalyzerCore -lclangStaticAnalyzerFrontend -lclangTooling -lclangToolingASTDiff -lclangToolingCore -lclangToolingInclusions -lclangToolingRefactor $(END_GROUP) -lstdc++

LLD_LIBS = $(START_GROUP) -llldCOFF -llldCommon -llldCore -llldDriver -llldELF -llldMachO -llldMinGW -llldReaderWriter -llldWasm -llldYAML $(END_GROUP)


# For static linking.
CGO_CPPFLAGS=$(shell $(LLVM_BUILDDIR)/bin/llvm-config --cppflags) -I$(abspath $(CLANG_SRC))/include -I$(abspath $(LLD_SRC))/include
CGO_CXXFLAGS=-std=c++11
CGO_LDFLAGS=-L$(LLVM_BUILDDIR)/lib $(CLANG_LIBS) $(LLD_LIBS) $(shell $(LLVM_BUILDDIR)/bin/llvm-config --ldflags --libs --system-libs $(LLVM_COMPONENTS))


clean:
	@rm -rf build

FMT_PATHS = ./*.go compiler interp ir loader src/device/arm src/examples src/machine src/os src/reflect src/runtime src/sync src/syscall
fmt:
	@gofmt -l -w $(FMT_PATHS)
fmt-check:
	@unformatted=$$(gofmt -l $(FMT_PATHS)); [ -z "$$unformatted" ] && exit 0; echo "Unformatted:"; for fn in $$unformatted; do echo "  $$fn"; done; exit 1

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


# Get LLVM sources.
llvm/README.txt:
	git clone -b release_80 https://github.com/llvm-mirror/llvm.git llvm
llvm/tools/clang/README.txt:
	git clone -b release_80 https://github.com/llvm-mirror/clang.git llvm/tools/clang
llvm/tools/lld/README.md:
	git clone -b release_80 https://github.com/llvm-mirror/lld.git llvm/tools/lld
llvm-source: llvm/README.txt llvm/tools/clang/README.txt llvm/tools/lld/README.md

# Configure LLVM.
llvm-build/build.ninja: llvm-source
	mkdir -p llvm-build; cd llvm-build; cmake -G Ninja ../llvm "-DLLVM_TARGETS_TO_BUILD=X86;ARM;AArch64;WebAssembly" "-DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD=AVR" -DCMAKE_BUILD_TYPE=Release -DLLVM_ENABLE_ASSERTIONS=OFF -DLIBCLANG_BUILD_STATIC=ON

# Build LLVM.
llvm-build: llvm-build/build.ninja
	cd llvm-build; ninja


# Build the Go compiler.
build/tinygo:
	@if [ ! -f llvm-build/bin/llvm-config ]; then echo "Fetch and build LLVM first by running:\n  make llvm-source\n  make llvm-build"; exit 1; fi
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" go build -o build/tinygo -tags byollvm .

test:
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" go test -v -tags byollvm .

release: build/tinygo gen-device
	@mkdir -p build/release/tinygo/bin
	@mkdir -p build/release/tinygo/lib/clang/include
	@mkdir -p build/release/tinygo/lib/CMSIS/CMSIS
	@mkdir -p build/release/tinygo/lib/compiler-rt/lib
	@mkdir -p build/release/tinygo/lib/nrfx
	@mkdir -p build/release/tinygo/pkg/armv6m-none-eabi
	@mkdir -p build/release/tinygo/pkg/armv7m-none-eabi
	@mkdir -p build/release/tinygo/pkg/armv7em-none-eabi
	@echo copying source files
	@cp -p  build/tinygo                 build/release/tinygo/bin
	@cp -p $(abspath $(CLANG_SRC))/lib/Headers/*.h build/release/tinygo/lib/clang/include
	@cp -rp lib/CMSIS/CMSIS/Include      build/release/tinygo/lib/CMSIS/CMSIS
	@cp -rp lib/CMSIS/README.md          build/release/tinygo/lib/CMSIS
	@cp -rp lib/compiler-rt/lib/builtins build/release/tinygo/lib/compiler-rt/lib
	@cp -rp lib/compiler-rt/LICENSE.TXT  build/release/tinygo/lib/compiler-rt
	@cp -rp lib/compiler-rt/README.txt   build/release/tinygo/lib/compiler-rt
	@cp -rp lib/nrfx/*                   build/release/tinygo/lib/nrfx
	@cp -rp src                          build/release/tinygo/src
	@cp -rp targets                      build/release/tinygo/targets
	./build/tinygo build-builtins -target=armv6m-none-eabi  -o build/release/tinygo/pkg/armv6m-none-eabi/compiler-rt.a
	./build/tinygo build-builtins -target=armv7m-none-eabi  -o build/release/tinygo/pkg/armv7m-none-eabi/compiler-rt.a
	./build/tinygo build-builtins -target=armv7em-none-eabi -o build/release/tinygo/pkg/armv7em-none-eabi/compiler-rt.a
	tar -czf build/release.tar.gz -C build/release tinygo
