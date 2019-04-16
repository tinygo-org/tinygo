# Building TinyGo

TinyGo depends on LLVM and libclang, which are both big C++ libraries. It can
also optionally use a built-in lld to ease cross compiling. There are two ways
these can be linked: dynamically and statically. The default is dynamic linking
because it is fast and works almost out of the box on Debian-based systems with
the right libraries installed.

This guide describes how to statically link TinyGo against LLVM, libclang and
lld so that the binary can be easily moved between systems. It also shows how to
build a release tarball that includes this binary and all necessary extra files.

## Dependencies

LLVM, Clang and LLD are quite light on dependencies, requiring only standard
build tools to be built. Go is of course necessary to build TinyGo itself.

  * Go (1.11+)
  * [dep](https://golang.github.io/dep/)
  * Standard build tools (gcc/clang)
  * git or subversion
  * CMake
  * [Ninja](https://ninja-build.org/) or make (preferably Ninja)

The rest of this guide assumes you're running Linux, but it should be equivalent
on a different system like Mac.

## Download the source

The first step is to get the source code. Place it in some directory, assuming
`$HOME/src` here, but you can pick a different one of course:

    git clone -b release_80 https://github.com/llvm-mirror/llvm.git $HOME/src/llvm
    git clone -b release_80 https://github.com/llvm-mirror/clang.git $HOME/src/llvm/tools/clang
    git clone -b release_80 https://github.com/llvm-mirror/lld.git $HOME/src/llvm/tools/lld
    go get -d github.com/tinygo-org/tinygo
    cd $HOME/go/src/github.com/tinygo-org/tinygo
    dep ensure -vendor-only # download dependencies

Note that Clang and LLD must be placed inside the tools subdirectory of LLVM to
be automatically built with the rest of the system.

## Build LLVM, Clang, LLD

Building LLVM is quite easy compared to some other software packages. However,
the default configuration is _not_ optimized for distribution. It is optimized
for development, meaning that binaries produce accurate error messages at the
cost of huge binaries and slow compiles.

Before configuring, you may want to set the following environment variables to
speed up the build. Most Linux distributions ship with GCC as the default
compiler, but Clang is significantly faster and uses much less memory while
producing binaries that are about as fast.

    export CC=clang
    export CXX=clang++

Make a build directory. LLVM requires out-of-tree builds:

    mkdir $HOME/src/llvm-build
    cd $HOME/src/llvm-build

Configure LLVM with CMake:

    cmake -G Ninja ../llvm "-DLLVM_TARGETS_TO_BUILD=X86;ARM;AArch64;WebAssembly" "-DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD=AVR" -DCMAKE_BUILD_TYPE=Release -DLLVM_ENABLE_ASSERTIONS=OFF -DLIBCLANG_BUILD_STATIC=ON

You can also choose a different build system than Ninja, but Ninja is fast.

There are various options you can tune here, but the options given above are
preferable for releases. Here is what they do:

  * `LLVM_TARGETS_TO_BUILD` and `LLVM_EXPERIMENTAL_TARGETS_TO_BUILD`: the
    targets that are natively supported by the LLVM code generators. The targets
    listed here are the ones supported by TinyGo. Note that LLVM is a cross
    compiler by default, unlike some other compilers.
  * `CMAKE_BUILD_TYPE`: the default is Debug, which produces large inefficient
    binaries that are easy to debug. We want small and fast binaries.
  * `LLVM_ENABLE_ASSERTIONS`: the default is ON, which greatly slows down LLVM
    and is only really useful during development. Disable them here.
  * `LIBCLANG_BUILD_STATIC`: unlike LLVM, libclang is built as a shared library
    by default. We want a static library for easy distribution.

Now build it:

    ninja # or make, if you choose make in the previous step

This can take over an hour depending on the speed of your system.

## Build TinyGo

Now that you have a working version of LLVM, build TinyGo using it. You need to
specify the directories to the LLVM build directory and to the Clang and LLD source.

    cd $HOME/go/src/github.com/tinygo-org/tinygo
    make LLVM_BUILDDIR=$HOME/src/llvm-build CLANG_SRC=$HOME/src/llvm/tools/clang LLD_SRC=$HOME/src/llvm/tools/lld

## Verify TinyGo

Try running TinyGo:

    ./build/tinygo help

Also, make sure the `tinygo` binary really is statically linked. Check this
using `ldd` (not to be confused with `lld`):

    ldd ./build/tinygo

The result should not contain libclang or libLLVM.

## Make a release tarball

Now that we have a working static build, it's time to make a release tarball.
This is just a slight change from the command to build TinyGo:

    cd $HOME/go/src/github.com/tinygo-org/tinygo
    make release LLVM_BUILDDIR=$HOME/src/llvm-build CLANG_SRC=$HOME/src/llvm/tools/clang LLD_SRC=$HOME/src/llvm/tools/lld

The release tarball is stored in build/release.tar.gz, and can be extracted with
the following command:

    tar -xvf path/to/release.tar.gz

TinyGo will get extracted to a `tinygo` directory. You can then call it with:

    ./tinygo/bin/tinygo
