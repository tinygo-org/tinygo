package loader

/*
#include <clang-c/Index.h> // if this fails, install libclang-7-dev

// The gateway function
int tinygo_clang_visitor_cgo(CXCursor c, CXCursor parent, CXClientData client_data) {
	int tinygo_clang_visitor(CXCursor c, CXCursor parent, CXClientData client_data);
	return tinygo_clang_visitor(c, parent, client_data);
}
*/
import "C"
