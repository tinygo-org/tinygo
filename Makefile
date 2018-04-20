
# aliases
all: tgo
tgo: build/tgo
test: build/hello.o

.PHONY: all tgo test run-test clean

CFLAGS = -Wall -Werror -O2 -g -flto

build/tgo: *.go
	@mkdir -p build
	@go build -o build/tgo -i .

build/hello.o: build/tgo src/examples/hello/*.go src/runtime/*.go
	@./build/tgo -printir -o build/hello.o examples/hello

build/runtime.o: src/runtime/*.c src/runtime/*.h
	clang $(CFLAGS) -c -o $@ src/runtime/*.c

build/hello: build/hello.o build/runtime.o
	@clang $(CFLAGS) -o $@ $^

run-test: build/hello
	@./build/hello

clean:
	@rm -rf build
