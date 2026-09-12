# Build the TinyGo compiler on top of a prebuilt LLVM image.
# Build the base image first with:
#   docker build -t tinygo-llvm-build -f Dockerfile.llvm .
# Or use the image that CI published:
#   docker build --build-arg LLVM_IMAGE=ghcr.io/tinygo-org/llvm-22:<tag> .
ARG LLVM_IMAGE=tinygo-llvm-build

# tinygo-compiler-build stage builds the compiler itself
FROM ${LLVM_IMAGE} AS tinygo-compiler-build

COPY . /tinygo

# build the compiler and tools
RUN cd /tinygo/ && \
    git submodule update --init && \
    make gen-device -j4 && \
    make build/release

# tinygo-compiler copies the compiler build over to a base Go container (without
# all the build tools etc).
FROM golang:1.27 AS tinygo-compiler

# Copy tinygo build.
COPY --from=tinygo-compiler-build /tinygo/build/release/tinygo /tinygo

# Configure the container.
ENV PATH="${PATH}:/tinygo/bin"
CMD ["tinygo"]
