# Build the release tarball and the Debian package.

build/release: tinygo gen-device $(if $(filter 1,$(USE_SYSTEM_BINARYEN)),,binaryen)
	@mkdir -p build/release/tinygo/bin
	@mkdir -p build/release/tinygo/lib/bdwgc
	@mkdir -p build/release/tinygo/lib/clang/include
	@mkdir -p build/release/tinygo/lib/CMSIS/CMSIS
	@mkdir -p build/release/tinygo/lib/macos-minimal-sdk
	@mkdir -p build/release/tinygo/lib/mingw-w64/mingw-w64-crt/crt
	@mkdir -p build/release/tinygo/lib/mingw-w64/mingw-w64-crt/math
	@mkdir -p build/release/tinygo/lib/mingw-w64/mingw-w64-crt/lib-common
	@mkdir -p build/release/tinygo/lib/mingw-w64/mingw-w64-headers/defaults
	@mkdir -p build/release/tinygo/lib/musl/arch
	@mkdir -p build/release/tinygo/lib/musl/crt
	@mkdir -p build/release/tinygo/lib/musl/src
	@mkdir -p build/release/tinygo/lib/nrfx
	@mkdir -p build/release/tinygo/lib/picolibc/newlib/libc
	@mkdir -p build/release/tinygo/lib/picolibc/newlib/libm
	@mkdir -p build/release/tinygo/lib/wasi-libc/dlmalloc
	@mkdir -p build/release/tinygo/lib/wasi-libc/libc-bottom-half
	@mkdir -p build/release/tinygo/lib/wasi-libc/libc-top-half/musl/arch
	@mkdir -p build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@mkdir -p build/release/tinygo/lib/wasi-cli/
	@echo copying source files
	@cp -p  build/tinygo$(EXE)           build/release/tinygo/bin
ifneq ($(USE_SYSTEM_BINARYEN),1)
	@cp -p  build/wasm-opt$(EXE)         build/release/tinygo/bin
endif
	@cp -rp lib/bdwgc/*                  build/release/tinygo/lib/bdwgc
	@cp -p $(abspath $(CLANG_SRC))/lib/Headers/*.h build/release/tinygo/lib/clang/include
	@cp -rp lib/CMSIS/CMSIS/Include      build/release/tinygo/lib/CMSIS/CMSIS
	@cp -rp lib/CMSIS/README.md          build/release/tinygo/lib/CMSIS
	@cp -rp lib/macos-minimal-sdk/*      build/release/tinygo/lib/macos-minimal-sdk
	@cp -rp lib/musl/arch/aarch64        build/release/tinygo/lib/musl/arch
	@cp -rp lib/musl/arch/arm            build/release/tinygo/lib/musl/arch
	@cp -rp lib/musl/arch/generic        build/release/tinygo/lib/musl/arch
	@cp -rp lib/musl/arch/i386           build/release/tinygo/lib/musl/arch
	@cp -rp lib/musl/arch/mips           build/release/tinygo/lib/musl/arch
	@cp -rp lib/musl/arch/x86_64         build/release/tinygo/lib/musl/arch
	@cp -rp lib/musl/crt/crt1.c          build/release/tinygo/lib/musl/crt
	@cp -rp lib/musl/COPYRIGHT           build/release/tinygo/lib/musl
	@cp -rp lib/musl/include             build/release/tinygo/lib/musl
	@cp -rp lib/musl/src/conf            build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/ctype           build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/env             build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/errno           build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/exit            build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/fcntl           build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/include         build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/internal        build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/legacy          build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/locale          build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/linux           build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/malloc          build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/mman            build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/math            build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/misc            build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/multibyte       build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/sched           build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/signal          build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/stdio           build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/stdlib          build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/string          build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/thread          build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/time            build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/unistd          build/release/tinygo/lib/musl/src
	@cp -rp lib/musl/src/process         build/release/tinygo/lib/musl/src
	@cp -rp lib/mingw-w64/mingw-w64-crt/crt/pseudo-reloc.c          build/release/tinygo/lib/mingw-w64/mingw-w64-crt/crt
	@cp -rp lib/mingw-w64/mingw-w64-crt/def-include                 build/release/tinygo/lib/mingw-w64/mingw-w64-crt
	@cp -rp lib/mingw-w64/mingw-w64-crt/gdtoa                       build/release/tinygo/lib/mingw-w64/mingw-w64-crt
	@cp -rp lib/mingw-w64/mingw-w64-crt/include                     build/release/tinygo/lib/mingw-w64/mingw-w64-crt
	@cp -rp lib/mingw-w64/mingw-w64-crt/lib-common/api-ms-win-crt-* build/release/tinygo/lib/mingw-w64/mingw-w64-crt/lib-common
	@cp -rp lib/mingw-w64/mingw-w64-crt/lib-common/advapi32.def.in  build/release/tinygo/lib/mingw-w64/mingw-w64-crt/lib-common
	@cp -rp lib/mingw-w64/mingw-w64-crt/lib-common/kernel32.def.in  build/release/tinygo/lib/mingw-w64/mingw-w64-crt/lib-common
	@cp -rp lib/mingw-w64/mingw-w64-crt/lib-common/msvcrt.def.in    build/release/tinygo/lib/mingw-w64/mingw-w64-crt/lib-common
	@cp -rp lib/mingw-w64/mingw-w64-crt/math/x86                    build/release/tinygo/lib/mingw-w64/mingw-w64-crt/math
	@cp -rp lib/mingw-w64/mingw-w64-crt/misc                        build/release/tinygo/lib/mingw-w64/mingw-w64-crt
	@cp -rp lib/mingw-w64/mingw-w64-crt/stdio                       build/release/tinygo/lib/mingw-w64/mingw-w64-crt
	@cp -rp lib/mingw-w64/mingw-w64-headers/crt/                    build/release/tinygo/lib/mingw-w64/mingw-w64-headers
	@cp -rp lib/mingw-w64/mingw-w64-headers/defaults/include        build/release/tinygo/lib/mingw-w64/mingw-w64-headers/defaults
	@cp -rp lib/mingw-w64/mingw-w64-headers/include                 build/release/tinygo/lib/mingw-w64/mingw-w64-headers
	@cp -rp lib/nrfx/*                   build/release/tinygo/lib/nrfx
	@cp -rp lib/picolibc/newlib/libc/ctype       build/release/tinygo/lib/picolibc/newlib/libc
	@cp -rp lib/picolibc/newlib/libc/include     build/release/tinygo/lib/picolibc/newlib/libc
	@cp -rp lib/picolibc/newlib/libc/locale      build/release/tinygo/lib/picolibc/newlib/libc
	@cp -rp lib/picolibc/newlib/libc/string      build/release/tinygo/lib/picolibc/newlib/libc
	@cp -rp lib/picolibc/newlib/libc/tinystdio   build/release/tinygo/lib/picolibc/newlib/libc
	@cp -rp lib/picolibc/newlib/libm/common      build/release/tinygo/lib/picolibc/newlib/libm
	@cp -rp lib/picolibc/newlib/libm/math        build/release/tinygo/lib/picolibc/newlib/libm
	@cp -rp lib/picolibc-stdio.c         build/release/tinygo/lib
	@cp -rp lib/wasi-libc/dlmalloc/src                              build/release/tinygo/lib/wasi-libc/dlmalloc
	@cp -rp lib/wasi-libc/libc-bottom-half/cloudlibc                build/release/tinygo/lib/wasi-libc/libc-bottom-half
	@cp -rp lib/wasi-libc/libc-bottom-half/headers                  build/release/tinygo/lib/wasi-libc/libc-bottom-half
	@cp -rp lib/wasi-libc/libc-bottom-half/sources                  build/release/tinygo/lib/wasi-libc/libc-bottom-half
	@cp -rp lib/wasi-libc/libc-top-half/headers                     build/release/tinygo/lib/wasi-libc/libc-top-half
	@cp -rp lib/wasi-libc/libc-top-half/musl/arch/generic           build/release/tinygo/lib/wasi-libc/libc-top-half/musl/arch
	@cp -rp lib/wasi-libc/libc-top-half/musl/arch/wasm32            build/release/tinygo/lib/wasi-libc/libc-top-half/musl/arch
	@cp -rp lib/wasi-libc/libc-top-half/musl/include                build/release/tinygo/lib/wasi-libc/libc-top-half/musl
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/conf               build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/dirent             build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/env                build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/errno              build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/exit               build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/fcntl              build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/fenv               build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/include            build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/internal           build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/legacy             build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/locale             build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/math               build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/misc               build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/multibyte          build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/network            build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/stat               build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/stdio              build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/stdlib             build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/string             build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/thread             build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/time               build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/musl/src/unistd             build/release/tinygo/lib/wasi-libc/libc-top-half/musl/src
	@cp -rp lib/wasi-libc/libc-top-half/sources                     build/release/tinygo/lib/wasi-libc/libc-top-half
	@cp -rp lib/wasi-cli/wit                                        build/release/tinygo/lib/wasi-cli/wit
	@cp -rp ${LLVM_PROJECTDIR}/compiler-rt/lib/builtins build/release/tinygo/lib/compiler-rt-builtins
	@cp -rp ${LLVM_PROJECTDIR}/compiler-rt/LICENSE.TXT  build/release/tinygo/lib/compiler-rt-builtins
	@cp -rp src                          build/release/tinygo/src
	@cp -rp targets                      build/release/tinygo/targets

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
