# tinygo-llvm stage obtains the llvm source for TinyGo
FROM golang:1.22 AS tinygo-llvm

RUN apt-get update && \
    apt-get install -y apt-utils make cmake clang-15 ninja-build

COPY ./GNUmakefile /tinygo/GNUmakefile

RUN cd /tinygo/ && \
    make llvm-source

# tinygo-llvm-build stage build the custom llvm with xtensa support
FROM tinygo-llvm AS tinygo-llvm-build

RUN cd /tinygo/ && \
    make llvm-build

# tinygo-compiler-build stage builds the compiler itself
FROM tinygo-llvm-build AS tinygo-compiler-build

COPY . /tinygo

# build the compiler and tools
RUN cd /tinygo/ && \
    git submodule update --init --recursive && \
    make gen-device -j4 && \
    make build/release

# tinygo-compiler copies the compiler build over to a base Go container (without
# all the build tools etc).
FROM golang:1.21 AS tinygo-compiler

# Copy tinygo build.
COPY --from=tinygo-compiler-build /tinygo/build/release/tinygo /tinygo

# Configure the container.
ENV PATH="${PATH}:/tinygo/bin"
CMD ["tinygo"]
