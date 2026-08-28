# Build the TinyGo compiler, plus housekeeping and code generation helpers.

.PHONY: all tinygo clean fmt fmt-check

clean: ## Remove build directory
	@rm -rf build

FMT_PATHS = ./*.go builder cgo/*.go compiler interp loader src transform
fmt: ## Reformat source
	@gofmt -l -w $(FMT_PATHS)
fmt-check: ## Warn if any source needs reformatting
	@unformatted=$$(gofmt -l $(FMT_PATHS)); [ -z "$$unformatted" ] && exit 0; echo "Unformatted:"; for fn in $$unformatted; do echo "  $$fn"; done; exit 1


# Generate WASI syscall bindings
WASM_TOOLS_MODULE=go.bytecodealliance.org
.PHONY: wasi-syscall
wasi-syscall: wasi-cm
	rm -rf ./src/internal/wasi/*
	go run $(WASM_TOOLS_MODULE)/cmd/wit-bindgen-go generate --versioned -o ./src/internal -p internal --cm internal/cm ./lib/wasi-cli/wit

# Copy package cm into src/internal/cm
.PHONY: wasi-cm
wasi-cm:
	rm -rf ./src/internal/cm/*
	rsync -rv --delete --exclude go.mod --exclude '*_test.go' --exclude '*_json.go' --exclude '*.md' --exclude LICENSE $(shell go list -m -f {{.Dir}} $(WASM_TOOLS_MODULE)/cm)/ ./src/internal/cm

# Check for Node.js used during WASM tests.
MIN_NODEJS_VERSION=22

.PHONY: check-nodejs-version
check-nodejs-version:
	@# Check whether NodeJS is available.
	@if ! command -v node 2>&1 >/dev/null; then echo "Install NodeJS version ${MIN_NODEJS_VERSION}+ to run tests."; exit 1; fi

	@# Check whether the version is high enough.
	@if [ "`node -v | sed 's/v\([0-9]\+\).*/\\1/g'`" -lt $(MIN_NODEJS_VERSION) ]; then echo "Install NodeJS version $(MIN_NODEJS_VERSION)+ to run tests."; exit 1; fi

tinygo: ## Build the TinyGo compiler
	@if [ ! -f "$(LLVM_BUILDDIR)/bin/llvm-config" ]; then echo "Fetch and build LLVM first by running:"; echo "  $(MAKE) llvm-source"; echo "  $(MAKE) $(LLVM_BUILDDIR)"; exit 1; fi
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CFLAGS="$(CGO_CFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GOENVFLAGS) $(GO) build -buildmode exe -o build/tinygo$(EXE) -tags "byollvm llvm22 osusergo" .
