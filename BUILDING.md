# Building TinyGo

TinyGo depends on LLVM and libclang, which are both big C++ libraries. It can
also optionally use a built-in lld to ease cross compiling. There are two ways
these can be linked: dynamically and statically. An install with `go install` is
dynamic linking because it is fast and works almost out of the box on
Debian-based systems with the right packages installed.

This guide describes how to statically link TinyGo against LLVM, libclang and
lld so that the binary can be easily moved between systems. It also shows how to
build a release tarball that includes this binary and all necessary extra files.

**Note**: this documentation describes how to build a statically linked release
tarball. If you want to help with development of TinyGo itself, you should follow the guide located at https://tinygo.org/docs/guides/build/

## Dependencies

LLVM, Clang and LLD are quite light on dependencies, requiring only standard
build tools to be built. Go is of course necessary to build TinyGo itself.

  * Go (1.19+)
  * GNU Make
  * Standard build tools (gcc/clang)
  * git
  * CMake
  * [Ninja](https://ninja-build.org/)

The rest of this guide assumes you're running Linux, but it should be equivalent
on a different system like Mac.

## Using GNU Make

The static build of TinyGo is driven by GNUmakefile, which includes the topic
files in the `make/` directory (`config.mk`, `llvm.mk`, `gen-device.mk`,
`build.mk`, `test.mk`, `smoketest.mk`, `release.mk`, and `tools.mk`).
It provides a help target for quick reference:

    % make help
    clean                           Remove build directory
    fmt                             Reformat source
    fmt-check                       Warn if any source needs reformatting
    gen-device                      Generate microcontroller-specific sources
    llvm-source                     Get LLVM sources
    llvm-build                      Build LLVM
    tinygo                          Build the TinyGo compiler
    lint                            Lint source tree
    spell                           Spellcheck source tree

## Download the source

The first step is to download the TinyGo sources (use `--recursive` if you clone
the git repository). Then, inside the directory, download the LLVM source:

    make llvm-source

You can also store LLVM outside of the TinyGo root directory by setting the
`LLVM_BUILDDIR`, `CLANG_SRC` and `LLD_SRC` make variables, but that is not
covered by this guide.

## Build LLVM, Clang, LLD

Before starting the build, you may want to set the following environment
variables to speed up the build. Most Linux distributions ship with GCC as the
default compiler, but Clang is significantly faster and uses much less memory
while producing binaries that are about as fast.

    export CC=clang
    export CXX=clang++

`make/config.mk` holds a default configuration that is good for most users. It
builds a release version of LLVM (optimized, no asserts) and includes all
targets supported by TinyGo:

    make llvm-build

This can take over an hour depending on the speed of your system.

## Build TinyGo

The last step of course is to build TinyGo itself. This can again be done with
make:

    make

## Verify TinyGo

Try running TinyGo:

    ./build/tinygo help

Also, make sure the `tinygo` binary really is statically linked. The command to check for 
dynamic dependencies differs depending on your operating system.

On Linux, use `ldd` (not to be confused with `lld`):

    ldd ./build/tinygo

On macOS, use otool -L:

    otool -L ./build/tinygo

The result should not contain libclang or libLLVM.

## Make a release tarball

Now that we have a working static build, it's time to make a release tarball:

    make release

If you did not clone the repository with the `--recursive` option, you will get errors until you initialize the project submodules:

    git submodule update --init

The release tarball is stored in build/release.tar.gz, and can be extracted with
the following command (for example in ~/lib):

    tar -xvf path/to/release.tar.gz

TinyGo will get extracted to a `tinygo` directory. You can then call it with:

    ./tinygo/bin/tinygo

## Publish a release

The `Release` workflow (`.github/workflows/release.yml`) publishes releases. It
does not build anything. The Linux, macOS and Windows workflows already build
every file that a release needs when the `release` branch is pushed, so the
release workflow collects the artifacts of those runs for the tagged commit.
What ships is what was tested.

 1. On the `dev` branch, set `const version` in `goenv/version.go` to the new
    version (without a `v` prefix), and add the entry to `CHANGELOG.md`.
 2. Merge `dev` into the `release` branch.
 3. Tag that commit and push the tag:

        git tag v0.42.0
        git push origin v0.42.0

    The tag must be `v` plus the version in `goenv/version.go`, because the
    release file names come from that constant.
 4. The workflow waits for the Linux, macOS and Windows runs of the tagged
    commit, collects their nine files, and creates a **draft** release. The
    release notes come from the `CHANGELOG.md` entry for that version.
 5. Review the draft release and publish it.
 6. On the `dev` branch, set `goenv/version.go` to the next `-dev` version.

To release again after a failure, delete the draft release and start the
workflow from the Actions tab with the tag as its input.

GitHub keeps a SHA-256 digest of every published file. The digest is not shown
on the release page, but it can be printed with:

    gh release view v0.42.0 --json assets --jq '.assets[] | "\(.digest)  \(.name)"'
