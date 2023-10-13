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
# You can also do many other things from this environment. Building and flashing
# should work as you're used to: it's not a VM or container so there are no
# access restrictions and you're running in the same host environment - just
# with a slightly different set of tools available.
{
  inputs = {
    # Use a recent stable release, but fix the version to make it reproducible.
    # This version should be updated from time to time.
    nixpkgs.url = "nixpkgs/nixos-23.05";
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
            llvmPackages_16.llvm
            llvmPackages_16.libclang
            # Additional dependencies needed at runtime, for building and/or
            # flashing.
            llvmPackages_16.lld
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
            # Ugly hack to make the Clang resources directory available.
            # Perhaps there is a cleaner way to do it, but this works.
            export GOFLAGS="\"-ldflags=-X github.com/tinygo-org/tinygo/goenv.clangResourceDir=${llvmPackages_16.clang.cc.lib}/lib/clang/16"\"
          '';
        };
      }
    );
}
