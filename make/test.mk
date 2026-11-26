# Test suites: compiler tests, standard library tests, benchmarks, and the corpus.

.PHONY: test wasmtest golden-check

test: check-nodejs-version
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=1h -buildmode exe -tags "byollvm llvm22 osusergo" $(GOTESTPKGS)

golden-check: ## Fail if regenerating the golden IR files changes them
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test -buildmode exe -tags "byollvm llvm22 osusergo" ./compiler ./transform -update
	git diff --exit-code -- compiler/testdata transform/testdata

TEST_ADDITIONAL_FLAGS ?=
STDLIB_TEST_FLAGS ?=

stdlib-test = GOOS="$(shell $(GO) env GOHOSTOS)" GOARCH="$(shell $(GO) env GOHOSTARCH)" \
	$(GO) run ./tools/stdlibtest -goos="$(GOOS)" -goarch="$(GOARCH)" $(STDLIB_TEST_FLAGS) $(1) \
	"$(GO)" "$(TINYGO)" $(TEST_ADDITIONAL_FLAGS) $(2)
stdlib-test-sum = $(if $(filter -plan -packages,$(STDLIB_TEST_FLAGS)),,gotestsum --raw-command -- env) \
	$(call stdlib-test,-json -mode=$(1),-target=wasip2 -x -work)

.PHONY: tinygo-test tinygo-test-plan tinygo-test-fast tinygo-test-wasm
.PHONY: tinygo-test-wasi tinygo-test-wasip1 tinygo-test-wasip1-fast
.PHONY: tinygo-test-wasip2-slow tinygo-test-wasip2-fast tinygo-test-wasip2-sum-slow tinygo-test-wasip2-sum-fast
.PHONY: tinygo-test-wasip3-slow tinygo-test-wasip3-fast
.PHONY: tinygo-test-baremetal tinygo-bench tinygo-bench-fast
.PHONY: tinygo-bench-wasip1 tinygo-bench-wasip1-fast tinygo-bench-wasip2 tinygo-bench-wasip2-fast
.PHONY: report-stdlib-tests-pass

tinygo-test:
	$(call stdlib-test)
tinygo-test-plan:
	$(call stdlib-test,-plan)
tinygo-test-fast:
	$(call stdlib-test,-mode=fast)
tinygo-test-wasm:
	$(call stdlib-test,-mode=fast,-target=wasm)
tinygo-test-wasi:
	$(call stdlib-test,,-target=wasip1)
tinygo-test-wasip1:
	$(call stdlib-test,-goos=wasip1 -goarch=wasm)
tinygo-test-wasip1-fast:
	$(call stdlib-test,-mode=fast,-target=wasip1)
tinygo-test-wasip2-slow:
	$(call stdlib-test,-mode=slow,-target=wasip2)
tinygo-test-wasip2-fast:
	$(call stdlib-test,-mode=fast,-target=wasip2)
tinygo-test-wasip3-slow:
	$(call stdlib-test,-mode=slow,-target=wasip3)
tinygo-test-wasip3-fast:
	$(call stdlib-test,-mode=fast,-target=wasip3)
tinygo-test-wasip2-sum-slow:
	$(call stdlib-test-sum,slow)
tinygo-test-wasip2-sum-fast:
	$(call stdlib-test-sum,fast)
tinygo-test-baremetal:
	$(call stdlib-test,-mode=fast,-target=riscv-qemu)

tinygo-bench:
	$(call stdlib-test,-bench)
tinygo-bench-fast:
	$(call stdlib-test,-bench -mode=fast)
tinygo-bench-wasip1:
	$(call stdlib-test,-bench,-target=wasip1)
tinygo-bench-wasip1-fast:
	$(call stdlib-test,-bench -mode=fast,-target=wasip1)
tinygo-bench-wasip2:
	$(call stdlib-test,-bench,-target=wasip2)
tinygo-bench-wasip2-fast:
	$(call stdlib-test,-bench -mode=fast,-target=wasip2)

report-stdlib-tests-pass:
	@$(call stdlib-test,-report)

# Test external packages in a large corpus.
test-corpus:
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=1h -buildmode exe -tags "byollvm llvm22" -run TestCorpus . -corpus=testdata/corpus.yaml
test-corpus-fast:
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=1h -buildmode exe -tags "byollvm llvm22" -run TestCorpus -short . -corpus=testdata/corpus.yaml
test-corpus-wasi:
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=1h -buildmode exe -tags "byollvm llvm22" -run TestCorpus . -corpus=testdata/corpus.yaml -target=wasip1
test-corpus-wasip2:
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=1h -buildmode exe -tags "byollvm llvm22" -run TestCorpus . -corpus=testdata/corpus.yaml -target=wasip2

wasmtest:
	cd ./tests/wasm && $(GO) test .
