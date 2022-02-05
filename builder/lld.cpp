// +build byollvm

// This file provides C wrappers for liblld.

#include <lld/Common/Driver.h>

extern "C" {

bool tinygo_link_elf(int argc, char **argv) {
	std::vector<const char*> args(argv, argv + argc);
	return lld::elf::link(args, llvm::outs(), llvm::errs(), false, false);
}

bool tinygo_link_macho(int argc, char **argv) {
	std::vector<const char*> args(argv, argv + argc);
	return lld::macho::link(args, llvm::outs(), llvm::errs(), false, false);
}

bool tinygo_link_mingw(int argc, char **argv) {
	std::vector<const char*> args(argv, argv + argc);
	return lld::mingw::link(args, llvm::outs(), llvm::errs(), false, false);
}

bool tinygo_link_wasm(int argc, char **argv) {
	std::vector<const char*> args(argv, argv + argc);
	return lld::wasm::link(args, llvm::outs(), llvm::errs(), false, false);
}

} // external "C"
