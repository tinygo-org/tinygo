# TinyGo base stage installs Go 1.14, LLVM 10 and the TinyGo compiler itself.
FROM golang:1.14 AS tinygo-base

RUN wget -O- https://apt.llvm.org/llvm-snapshot.gpg.key| apt-key add - && \
    echo "deb http://apt.llvm.org/buster/ llvm-toolchain-buster-10 main" >> /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y llvm-10-dev libclang-10-dev git

COPY . /tinygo

# remove submodules directories and re-init them to fix any hard-coded paths
# after copying the tinygo directory in the previous step.
RUN cd /tinygo/ && \
    rm -rf ./lib/* && \
    git submodule sync && \
    git submodule update --init --recursive --force

COPY ./lib/picolibc-include/* /tinygo/lib/picolibc-include/

RUN cd /tinygo/ && \
    go install /tinygo/

# tinygo-wasm stage installs the needed dependencies to compile TinyGo programs for WASM.
FROM tinygo-base AS tinygo-wasm

COPY --from=tinygo-base /go/bin/tinygo /go/bin/tinygo
COPY --from=tinygo-base /tinygo/src /tinygo/src
COPY --from=tinygo-base /tinygo/targets /tinygo/targets

RUN wget -O- https://apt.llvm.org/llvm-snapshot.gpg.key| apt-key add - && \
    echo "deb http://apt.llvm.org/buster/ llvm-toolchain-buster-10 main" >> /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y libllvm10 lld-10

# tinygo-avr stage installs the needed dependencies to compile TinyGo programs for AVR microcontrollers.
FROM tinygo-base AS tinygo-avr

COPY --from=tinygo-base /go/bin/tinygo /go/bin/tinygo
COPY --from=tinygo-base /tinygo/src /tinygo/src
COPY --from=tinygo-base /tinygo/targets /tinygo/targets
COPY --from=tinygo-base /tinygo/Makefile /tinygo/
COPY --from=tinygo-base /tinygo/tools /tinygo/tools
COPY --from=tinygo-base /tinygo/lib /tinygo/lib

RUN cd /tinygo/ && \
    apt-get update && \
    apt-get install -y apt-utils make binutils-avr gcc-avr avr-libc && \
    make gen-device-avr && \
    apt-get autoremove -y && \
    apt-get clean

# tinygo-arm stage installs the needed dependencies to compile TinyGo programs for ARM microcontrollers.
FROM tinygo-base AS tinygo-arm

COPY --from=tinygo-base /go/bin/tinygo /go/bin/tinygo
COPY --from=tinygo-base /tinygo/src /tinygo/src
COPY --from=tinygo-base /tinygo/targets /tinygo/targets
COPY --from=tinygo-base /tinygo/Makefile /tinygo/
COPY --from=tinygo-base /tinygo/tools /tinygo/tools
COPY --from=tinygo-base /tinygo/lib /tinygo/lib

RUN cd /tinygo/ && \
    apt-get update && \
    apt-get install -y apt-utils make clang-10 && \
    make gen-device-nrf && make gen-device-stm32

# tinygo-all stage installs the needed dependencies to compile TinyGo programs for all platforms.
FROM tinygo-wasm AS tinygo-all

COPY --from=tinygo-base /tinygo/Makefile /tinygo/
COPY --from=tinygo-base /tinygo/tools /tinygo/tools
COPY --from=tinygo-base /tinygo/lib /tinygo/lib

RUN cd /tinygo/ && \
    apt-get update && \
    apt-get install -y apt-utils make clang-10 binutils-avr gcc-avr avr-libc && \
    make gen-device

CMD ["tinygo"]
