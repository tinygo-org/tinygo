# A Nix flake file, mainly intended for developing TinyGo.
# You can download Nix here, for use on your Linux or macOS system:
#   https://nixos.org/download.html
# After you have installed Nix, you can enter the development environment as
# follows:
#
#   nix develop
#
# This drops you into a bash shell, where you can install TinyGo simply using
# the following command:
#
#   go install
#
# That's all! Assuming you've set up your $PATH correctly, you can now use the
# tinygo command as usual:
#
#   tinygo version
#
# But you'll need a bit more to make TinyGo actually able to compile code:
#
#   make llvm-source            # fetch compiler-rt
#   git submodule update --init # fetch lots of other libraries and SVD files
#   make gen-device -j4         # build src/device/*/*.go files
#   make wasi-libc              # build support for wasi/wasm
#
# With this, you should have an environment that can compile anything - except
# for the Xtensa architecture (ESP8266/ESP32) because support for that lives in
# a separate LLVM fork.
#
# You can also do many other things from this environment. Building and flashing
# should work as you're used to: it's not a VM or container so there are no
# access restrictions and you're running in the same host environment - just
# with a slightly different set of tools available.
{
  inputs = {
    # Use a recent stable release, but fix the version to make it reproducible.
    # This version should be updated from time to time.
    nixpkgs.url = "nixpkgs/nixos-23.11";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      with pkgs;
      {
        devShells.default = mkShell {
          buildInputs = [
            # These dependencies are required for building tinygo (go install).
            go
            llvmPackages_17.llvm
            llvmPackages_17.libclang
            # Additional dependencies needed at runtime, for building and/or
            # flashing.
            llvmPackages_17.lld
            avrdude
            binaryen
            # Additional dependencies needed for on-chip debugging.
            # These tools are rather big (especially GDB) and not frequently
            # used, so are commented out. On-chip debugging is still possible if
            # these tools are available in the host environment.
            #gdb
            #openocd
          ];
          shellHook= ''
            # Configure CLANG, LLVM_AR, and LLVM_NM for `make wasi-libc`.
            # Without setting these explicitly, Homebrew versions might be used
            # or the default `ar` and `nm` tools might be used (which don't
            # support wasi).
            export CLANG="clang-17 -resource-dir ${llvmPackages_17.clang.cc.lib}/lib/clang/17"
            export LLVM_AR=llvm-ar
            export LLVM_NM=llvm-nm

            # Make `make smoketest` work (the default is `md5`, while Nix only
            # has `md5sum`).
            export MD5SUM=md5sum

            # Ugly hack to make the Clang resources directory available.
            export GOFLAGS="\"-ldflags=-X github.com/tinygo-org/tinygo/goenv.clangResourceDir=${llvmPackages_17.clang.cc.lib}/lib/clang/17"\"
          '';
        };
      }
    );
}
