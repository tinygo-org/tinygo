
all: tgo
tgo: build/tgo
test: build/hello.o

CFLAGS = -Wall -Werror -O2 -g -flto

.PHONY: all tgo test run-test clean

build/tgo: *.go
	@mkdir -p build
	@go build -o build/tgo -i .

build/hello.o: build/tgo hello/hello.go
	@./build/tgo -printir -o build/hello.o hello/hello.go

build/runtime.o: runtime/*.c runtime/*.h
	clang $(CFLAGS) -c -o $@ runtime/*.c

build/hello: build/hello.o build/runtime.o
	@clang $(CFLAGS) -o $@ $^

run-test: build/hello
	@./build/hello

clean:
	@rm -rf build
