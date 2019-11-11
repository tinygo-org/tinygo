// +build byollvm

// This file provides C wrappers for liblld.

#include <lld/Common/Driver.h>

extern "C" {

bool tinygo_link_elf(int argc, char **argv) {
	std::vector<const char*> args(argv, argv + argc);
	return lld::elf::link(args, false);
}

bool tinygo_link_wasm(int argc, char **argv) {
	std::vector<const char*> args(argv, argv + argc);
	return lld::wasm::link(args, false);
}

} // external "C"
