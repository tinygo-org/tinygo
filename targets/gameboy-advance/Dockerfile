# This dockerfile is intended to be built from the root of the tinygo source tree
# using a command like the following:
#
#   docker build --no-cache --rm -f targets/gameboy-advance/Dockerfile -t tinygo/tinygo_gba:latest .

# Currently Debian "sid" has llvm-8 required by TinyGo
FROM debian:sid AS tinygo-gba-base

# Install dependencies
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get -y update
RUN apt-get -y install wget bzip2 xz-utils libxml2 git gpg
RUN apt-get -y install make gcc g++ clang-8 libclang-8-dev lld-8 binutils-arm-none-eabi

# Install Go
WORKDIR /go
RUN wget https://dl.google.com/go/go1.12.7.linux-amd64.tar.gz
RUN tar -C / -xzvf go*.tar.gz
ENV PATH="/go/bin:${PATH}"

# Install TinyGo
WORKDIR /tinygo
ADD . .
RUN git submodule deinit --force --all
RUN go mod download
RUN go build -o bin/tinygo
RUN go build -o bin/tileprep ./tools/tileprep/
ENV PATH="/tinygo/bin:${PATH}"

WORKDIR /skel
ADD targets/gameboy-advance/Makefile .

WORKDIR /
ADD targets/gameboy-advance/run.sh .

WORKDIR /gba
ENTRYPOINT [ "/run.sh" ]
