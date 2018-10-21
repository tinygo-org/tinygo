FROM golang:latest

RUN wget -O- https://apt.llvm.org/llvm-snapshot.gpg.key| apt-key add - && \
    echo "deb http://apt.llvm.org/stretch/ llvm-toolchain-stretch-7 main" >> /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y llvm-7-dev

RUN wget -O- https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

COPY . /go/src/github.com/aykevl/tinygo

RUN cd /go/src/github.com/aykevl/tinygo/ && \
    dep ensure --vendor-only && \
    go install /go/src/github.com/aykevl/tinygo/

FROM golang:latest

COPY --from=0 /go/bin/tinygo /go/bin/tinygo

RUN wget -O- https://apt.llvm.org/llvm-snapshot.gpg.key| apt-key add - && \
    echo "deb http://apt.llvm.org/stretch/ llvm-toolchain-stretch-7 main" >> /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y libllvm7

ENTRYPOINT ["/go/bin/tinygo"]
