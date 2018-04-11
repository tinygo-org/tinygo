
all: tgo
tgo: build/tgo
test: build/hello.o

.PHONY: all tgo test test-run clean

build/tgo: *.go
	@mkdir -p build
	@go build -o build/tgo -i .

build/hello.o: build/tgo hello/hello.go
	@./build/tgo -printir -target x86_64-pc-linux-gnu -o build/hello.o hello/hello.go

build/hello: build/hello.o
	@clang -o build/hello build/hello.o

test-run: build/hello
	@./build/hello

clean:
	@rm -rf build
