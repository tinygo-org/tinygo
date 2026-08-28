# LLVM and Binaryen: component lists, link flags, source checkout, and build.

.PHONY: llvm-source $(LLVM_BUILDDIR)

LLVM_COMPONENTS = all-targets analysis asmparser asmprinter bitreader bitwriter codegen core coroutines coverage debuginfodwarf debuginfopdb dtlto executionengine frontenddriver frontendhlsl frontendopenmp instrumentation interpreter ipo irreader libdriver linker lto mc mcjit objcarcopts option profiledata scalaropts support target windowsdriver windowsmanifest

ifeq ($(OS),Windows_NT)
    EXE = .exe
    START_GROUP = -Wl,--start-group
    END_GROUP = -Wl,--end-group

    # PIC needs to be disabled for libclang to work.
    LLVM_OPTION += -DLLVM_ENABLE_PIC=OFF
    # Statically link the C++ and GCC runtime into LLVM tools so they don't
    # depend on MinGW DLLs that may not be on PATH when executed during the build.
    LLVM_OPTION += '-DCMAKE_EXE_LINKER_FLAGS=-static-libgcc -static-libstdc++'

    CGO_CPPFLAGS += -DCINDEX_NO_EXPORTS
    CGO_LDFLAGS += -static -static-libgcc -static-libstdc++
    CGO_LDFLAGS_EXTRA += -lversion

    USE_SYSTEM_BINARYEN ?= 1

else ifeq ($(uname),Darwin)
    MD5SUM ?= md5

    CGO_LDFLAGS += -lxar

    USE_SYSTEM_BINARYEN ?= 1

else ifeq ($(uname),FreeBSD)
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
CLANG_LIB_NAMES = clangAnalysis clangAnalysisLifetimeSafety clangAPINotes clangAST clangASTMatchers clangBasic clangCodeGen clangCrossTU clangDriver clangDynamicASTMatchers clangEdit clangExtractAPI clangFormat clangFrontend clangFrontendTool clangHandleCXX clangHandleLLVM clangIndex clangInstallAPI clangLex clangOptions clangParse clangRewrite clangRewriteFrontend clangSema clangSerialization clangSupport clangTooling clangToolingASTDiff clangToolingCore clangToolingInclusions
CLANG_LIBS = $(START_GROUP) $(addprefix -l,$(CLANG_LIB_NAMES)) $(END_GROUP) -lstdc++

# Libraries that should be linked in for the statically linked LLD.
LLD_LIB_NAMES = lldCOFF lldCommon lldELF lldMachO lldMinGW lldWasm
LLD_LIBS = $(START_GROUP) $(addprefix -l,$(LLD_LIB_NAMES)) $(END_GROUP)

# Other libraries that are needed to link TinyGo.
EXTRA_LIB_NAMES = LLVMDTLTO LLVMInterpreter LLVMMCA LLVMRISCVTargetMCA LLVMX86TargetMCA

# All libraries to be built and linked with the tinygo binary (lib/lib*.a).
LIB_NAMES = clang $(CLANG_LIB_NAMES) $(LLD_LIB_NAMES) $(EXTRA_LIB_NAMES)

# These build targets appear to be the only ones necessary to build all TinyGo
# dependencies. Only building a subset significantly speeds up rebuilding LLVM.
# The Makefile rules convert a name like lldELF to lib/liblldELF.a to match the
# library path (for ninja).
# This list also includes a few tools that are necessary as part of the full
# TinyGo build.
NINJA_BUILD_TARGETS = clang llvm-config llvm-ar llvm-nm lld $(addprefix lib/lib,$(addsuffix .a,$(LIB_NAMES)))

# For static linking.
ifneq ("$(wildcard $(LLVM_BUILDDIR)/bin/llvm-config*)","")
    CGO_CPPFLAGS+=$(shell $(LLVM_CONFIG_PREFIX) $(LLVM_BUILDDIR)/bin/llvm-config --cppflags) -I$(abspath $(LLVM_BUILDDIR))/tools/clang/include -I$(abspath $(CLANG_SRC))/include -I$(abspath $(LLD_SRC))/include
    CGO_CXXFLAGS=-std=c++17
ifneq ($(uname),Windows_NT)
    # Disable GCC DWARF compression: lld built without zlib cannot link
    # object files with ELFCOMPRESS_ZLIB debug sections.
    CGO_CFLAGS+=-gz=none
    CGO_CXXFLAGS+=-gz=none
endif
    CGO_LDFLAGS+=-L$(abspath $(LLVM_BUILDDIR)/lib) -lclang $(CLANG_LIBS) $(LLD_LIBS) $(shell $(LLVM_CONFIG_PREFIX) $(LLVM_BUILDDIR)/bin/llvm-config --ldflags --libs --system-libs $(LLVM_COMPONENTS)) -lstdc++ $(CGO_LDFLAGS_EXTRA)
endif

$(LLVM_PROJECTDIR)/llvm:
	git clone -b tinygo_22.x --depth=1 https://github.com/tinygo-org/llvm-project $(LLVM_PROJECTDIR)
llvm-source: $(LLVM_PROJECTDIR)/llvm ## Get LLVM sources

# Configure LLVM.
TINYGO_SOURCE_DIR=$(shell pwd)
$(LLVM_BUILDDIR)/build.ninja:
	mkdir -p $(LLVM_BUILDDIR) && cd $(LLVM_BUILDDIR) && cmake -G Ninja $(TINYGO_SOURCE_DIR)/$(LLVM_PROJECTDIR)/llvm "-DLLVM_TARGETS_TO_BUILD=X86;ARM;AArch64;AVR;Mips;RISCV;WebAssembly" "-DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD=Xtensa" -DCMAKE_BUILD_TYPE=Release -DLIBCLANG_BUILD_STATIC=ON -DLLVM_ENABLE_TERMINFO=OFF -DLLVM_ENABLE_ZLIB=OFF -DLLVM_ENABLE_ZSTD=OFF -DLLVM_ENABLE_LIBEDIT=OFF -DLLVM_ENABLE_Z3_SOLVER=OFF -DLLVM_ENABLE_OCAMLDOC=OFF -DLLVM_ENABLE_LIBXML2=OFF -DLLVM_ENABLE_PROJECTS="clang;lld" -DLLVM_TOOL_CLANG_TOOLS_EXTRA_BUILD=OFF -DCLANG_ENABLE_STATIC_ANALYZER=OFF -DCLANG_ENABLE_ARCMT=OFF $(LLVM_OPTION)

$(LLVM_BUILDDIR): $(LLVM_BUILDDIR)/build.ninja ## Build LLVM
	cd $(LLVM_BUILDDIR) && ninja $(NINJA_BUILD_TARGETS)

ifneq ($(USE_SYSTEM_BINARYEN),1)
# Build Binaryen
.PHONY: binaryen
binaryen: build/wasm-opt$(EXE)
build/wasm-opt$(EXE):
	mkdir -p build
	cd lib/binaryen && cmake -G Ninja . -DBUILD_STATIC_LIB=ON -DBUILD_TESTS=OFF -DENABLE_WERROR=OFF $(BINARYEN_OPTION) && ninja bin/wasm-opt$(EXE)
	cp lib/binaryen/bin/wasm-opt$(EXE) build/wasm-opt$(EXE)
endif
