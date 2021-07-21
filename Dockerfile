# TinyGo base stage installs the most recent Go 1.17.x, LLVM 11 and the TinyGo compiler itself.
FROM golang:1.17 AS tinygo-base

RUN wget -O- https://apt.llvm.org/llvm-snapshot.gpg.key| apt-key add - && \
    echo "deb http://apt.llvm.org/bullseye/ llvm-toolchain-bullseye-11 main" >> /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y llvm-11-dev libclang-11-dev lld-11 git

COPY . /tinygo

# remove submodules directories and re-init them to fix any hard-coded paths
# after copying the tinygo directory in the previous step.
RUN cd /tinygo/ && \
    rm -rf ./lib/* && \
    git submodule sync && \
    git submodule update --init --recursive --force

COPY ./lib/picolibc-stdio.c /tinygo/lib/picolibc-stdio.c

RUN cd /tinygo/ && \
    go install /tinygo/

# tinygo-wasm stage installs the needed dependencies to compile TinyGo programs for WASM.
FROM tinygo-base AS tinygo-wasm

COPY --from=tinygo-base /go/bin/tinygo /go/bin/tinygo
COPY --from=tinygo-base /tinygo/src /tinygo/src
COPY --from=tinygo-base /tinygo/targets /tinygo/targets

RUN cd /tinygo/ && \
    apt-get update && \
    apt-get install -y make clang-11 libllvm11 lld-11 && \
    make wasi-libc binaryen

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
    apt-get install -y apt-utils make clang-11 && \
    make gen-device-nrf && make gen-device-stm32

# tinygo-all stage installs the needed dependencies to compile TinyGo programs for all platforms.
FROM tinygo-wasm AS tinygo-all

COPY --from=tinygo-base /tinygo/Makefile /tinygo/
COPY --from=tinygo-base /tinygo/tools /tinygo/tools
COPY --from=tinygo-base /tinygo/lib /tinygo/lib

RUN cd /tinygo/ && \
    apt-get update && \
    apt-get install -y apt-utils make clang-11 binutils-avr gcc-avr avr-libc && \
    make gen-device

CMD ["tinygo"]
