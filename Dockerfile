# tinygo-llvm stage obtains the llvm source for TinyGo
FROM golang:1.17 AS tinygo-llvm

RUN apt-get update && \
    apt-get install -y apt-utils make cmake clang-11 binutils-avr gcc-avr avr-libc ninja-build

COPY ./Makefile /tinygo/Makefile

RUN cd /tinygo/ && \
    make llvm-source

# tinygo-llvm-build stage build the custom llvm with xtensa support
FROM tinygo-llvm AS tinygo-llvm-build

RUN cd /tinygo/ && \
    make llvm-build

# tinygo-xtensa stage installs tools needed for ESP32
FROM tinygo-llvm-build AS tinygo-xtensa

ARG xtensa_version="1.22.0-80-g6c4433a-5.2.0"
RUN cd /tmp/ && \
    wget -q https://dl.espressif.com/dl/xtensa-esp32-elf-linux64-${xtensa_version}.tar.gz && \
    tar xzf xtensa-esp32-elf-linux64-${xtensa_version}.tar.gz && \
    cp ./xtensa-esp32-elf/bin/xtensa-esp32-elf-ld /usr/local/bin/ && \
    rm -rf /tmp/xtensa*

# tinygo-compiler stage builds the compiler itself
FROM tinygo-xtensa AS tinygo-compiler

COPY . /tinygo

# update submodules
RUN cd /tinygo/ && \
    rm -rf ./lib/*/ && \
    git submodule sync && \
    git submodule update --init --recursive --force

RUN cd /tinygo/ && \
    make

# tinygo-tools stage installs the needed dependencies to compile TinyGo programs for all platforms.
FROM tinygo-compiler AS tinygo-tools

RUN cd /tinygo/ && \
    make wasi-libc binaryen && \
    make gen-device -j4 && \
    cp build/* $GOPATH/bin/

CMD ["tinygo"]
