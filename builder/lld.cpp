//go:build byollvm

// This file provides C wrappers for liblld.

#include <lld/Common/Driver.h>
#include <llvm/Support/Parallel.h>

LLD_HAS_DRIVER(coff)
LLD_HAS_DRIVER(elf)
LLD_HAS_DRIVER(mingw)
LLD_HAS_DRIVER(macho)
LLD_HAS_DRIVER(wasm)

static void configure() {
#if _WIN64
	// This is a hack to work around a hang in the LLD linker on Windows, with
	// -DLLVM_ENABLE_THREADS=ON. It has a similar effect as the -threads=1
	// linker flag, but with support for the COFF linker.
	llvm::parallel::strategy = llvm::hardware_concurrency(1);
#endif
}

extern "C" {

bool tinygo_link(int argc, char **argv) {
	configure();
	std::vector<const char*> args(argv, argv + argc);
	lld::Result r = lld::lldMain(args, llvm::outs(), llvm::errs(), LLD_ALL_DRIVERS);
	return !r.retCode;
}

} // external "C"
