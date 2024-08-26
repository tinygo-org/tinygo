# Command that's part of sizediff.yml. This is put in a separate file so that it
# still works after checking out the dev branch (that is, when going from LLVM
# 16 to LLVM 17 for example, both Clang 16 and Clang 17 are installed).

echo 'deb https://apt.llvm.org/noble/ llvm-toolchain-noble-19 main' | sudo tee /etc/apt/sources.list.d/llvm.list
wget -O - https://apt.llvm.org/llvm-snapshot.gpg.key | sudo apt-key add -
sudo apt-get update
sudo apt-get install --no-install-recommends -y \
    llvm-19-dev \
    clang-19 \
    libclang-19-dev \
    lld-19
