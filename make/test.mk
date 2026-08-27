# Test suites: compiler tests, standard library tests, benchmarks, and the corpus.

.PHONY: test wasmtest

test: check-nodejs-version
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=1h -buildmode exe -tags "byollvm llvm22 osusergo" $(GOTESTPKGS)

# Standard library packages that pass tests on darwin, linux, wasi, and windows, but take over a minute in wasi
TEST_PACKAGES_SLOW = \
	compress/bzip2 \
	crypto/dsa \
	index/suffixarray \

# Standard library packages that pass tests quickly on darwin, linux, wasi, and windows
TEST_PACKAGES_FAST = \
	cmp \
	compress/lzw \
	compress/zlib \
	container/heap \
	container/list \
	container/ring \
	crypto/ecdsa \
	crypto/elliptic \
	crypto/md5 \
	crypto/sha1 \
	crypto/sha256 \
	crypto/sha512 \
	database/sql/driver \
	debug/macho \
	embed/internal/embedtest \
	encoding \
	encoding/ascii85 \
	errors \
	encoding/asn1 \
	encoding/base32 \
	encoding/base64 \
	encoding/csv \
	encoding/hex \
	expvar \
	go/ast \
	go/format \
	go/scanner \
	go/token \
	go/version \
	hash \
	hash/adler32 \
	hash/crc64 \
	hash/fnv \
	html \
	internal/itoa \
	internal/profile \
	math \
	math/cmplx \
	net/http/internal/ascii \
	net/mail \
	net/url \
	os \
	path \
	reflect \
	sync \
	testing \
	testing/iotest \
	text/scanner \
	unicode \
	unicode/utf16 \
	unicode/utf8 \
	unique \
	$(nil)

# archive/zip requires os.ReadAt, which is not yet supported on windows
# bytes requires mmap
# compress/flate appears to hang on wasi
# crypto/aes needs reflect.Type.Method(), not yet implemented
# crypto/des fails on wasi, needs panic()/recover()
# crypto/hmac fails on wasi, it exits with a "slice out of range" panic
# debug/plan9obj requires os.ReadAt, which is not yet supported on windows
# encoding/xml takes a minute on linux and gives a stack overflow on wasi
# image fails on wasi, needs panic()/recover()
# io/ioutil requires os.ReadDir, which is not yet supported on windows or wasi
# mime: fails on wasi, needs panic()/recover()
# mime/multipart: needs wasip1 syscall.FDFLAG_NONBLOCK
# mime/quotedprintable requires syscall.Faccessat
# net/mail: needs wasip1  syscall.FDFLAG_NONBLOCK
# net/ntextproto: needs wasip1 syscall.FDFLAG_NONBLOCK
# regexp/syntax: fails on wasip1, needs panic()/recover()
# strconv: fails on wasi, needs panic()/recover()
# text/tabwriter: fails on wasi, needs panic()/recover()
# text/template/parse: fails on wasi, needs panic()/recover()
# testing/fstest requires os.ReadDir, which is not yet supported on windows or wasi

# Additional standard library packages that pass tests on individual platforms
TEST_PACKAGES_LINUX := \
	archive/zip \
	compress/flate \
	context \
	crypto/aes \
	crypto/des \
	crypto/ecdh \
	crypto/hmac \
	debug/dwarf \
	debug/plan9obj \
	encoding/xml \
	image \
	io/ioutil \
	mime \
	mime/multipart \
	mime/quotedprintable \
	net \
	net/mail \
	net/textproto \
	os/user \
	regexp/syntax \
	strconv \
	testing/fstest \
	text/tabwriter \
	text/template/parse

TEST_PACKAGES_DARWIN := $(TEST_PACKAGES_LINUX)

# os/user requires t.Skip() support
TEST_PACKAGES_WINDOWS := \
	compress/flate \
	crypto/des \
	crypto/hmac \
	image \
	mime \
	regexp/syntax \
	strconv \
	text/tabwriter \
	text/template/parse \
	$(nil)


# These packages cannot be tested on wasm, mostly because these tests assume a
# working filesystem. This could perhaps be fixed, by supporting filesystem
# access when running inside Node.js.
TEST_PACKAGES_WASM = $(filter-out $(TEST_PACKAGES_NONWASM), $(TEST_PACKAGES_FAST))
TEST_PACKAGES_NONWASM = \
	compress/lzw \
	compress/zlib \
	crypto/ecdsa \
	debug/macho \
	embed/internal/embedtest \
	expvar \
	go/format \
	os \
	testing \
	$(nil)

# These packages cannot be tested on baremetal.
#
# Some reasons why the tests don't pass on baremetal:
#
#   * No filesystem is available, so packages like compress/zlib can't be tested
#     (just like wasm).
#   * picolibc math functions apparently are less precise, the math package
#     fails on baremetal.
#   * Since Go 1.27 the crypto tests below go through cryptotest.TestHash, which
#     calls cryptotest.BoundarySlices. These targets report GOOS=linux, so they
#     build boundary.go (//go:build linux || darwin) rather than
#     boundary_compat.go, and that needs a working syscall.Mmap/syscall.Mprotect
#     which we don't have. See #5593.
TEST_PACKAGES_BAREMETAL = $(filter-out $(TEST_PACKAGES_NONBAREMETAL), $(TEST_PACKAGES_FAST))
TEST_PACKAGES_NONBAREMETAL = \
	$(TEST_PACKAGES_NONWASM) \
	$(TEST_PACKAGES_NOBOUNDARYSLICES) \
	math \
	$(nil)

TEST_PACKAGES_FAST_WASI = $(filter-out $(TEST_PACKAGES_NOWASI), $(TEST_PACKAGES_FAST))
TEST_PACKAGES_NOWASI = \
	crypto/ecdsa \
	$(nil)

# wasip1 reports GOOS=wasip1 and so gets the boundary_compat.go fallback, but
# wasip2 reports GOOS=linux and hits the same BoundarySlices problem as
# baremetal. On wasip2 syscall.Mmap returns ENOSYS and t.Fatalf cannot Goexit,
# so the test falls through and panics with "slice out of range".
TEST_PACKAGES_FAST_WASIP2 = $(filter-out $(TEST_PACKAGES_NOBOUNDARYSLICES), $(TEST_PACKAGES_FAST_WASI))

TEST_PACKAGES_NOBOUNDARYSLICES = \
	crypto/md5 \
	crypto/sha1 \
	crypto/sha256 \
	crypto/sha512 \
	$(nil)

# Report platforms on which each standard library package is known to pass tests
report-stdlib-tests-pass:
	$(eval jointmp := $(shell echo /tmp/join.$$$$))
	@for t in $(TEST_PACKAGES_DARWIN); do echo "$$t darwin"; done | sort > $(jointmp).darwin
	@for t in $(TEST_PACKAGES_LINUX); do echo "$$t linux"; done | sort > $(jointmp).linux
	@for t in $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_SLOW); do echo "$$t darwin linux wasi windows"; done | sort > $(jointmp).portable
	@join -a1 -a2 $(jointmp).darwin $(jointmp).linux | \
	join -a1 -a2 - $(jointmp).portable
	@rm $(jointmp).*

# Standard library packages that pass tests quickly on the current platform
ifeq ($(uname),Darwin)
TEST_PACKAGES_HOST := $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_DARWIN)
TEST_IOFS := true
TEST_ENCODING_XML := true
endif
ifeq ($(uname),Linux)
TEST_PACKAGES_HOST := $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_LINUX)
TEST_IOFS := true
TEST_ENCODING_XML := true
endif
ifeq ($(OS),Windows_NT)
TEST_PACKAGES_HOST := $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_WINDOWS)
TEST_IOFS := false
endif

TEST_SKIP_FLAG := -skip='TestExtraMethods|TestParseAndBytesRoundTrip/P256/Generic|TestAsValidation|TestUnmarshalNestingLimitSlice|TestUnmarshalNestingLimitStruct'
TEST_ADDITIONAL_FLAGS ?=

# Test known-working standard library packages.
# TODO: parallelize, and only show failing tests (no implied -v flag).
.PHONY: tinygo-test
tinygo-test:
	@# TestExtraMethods: used by many crypto packages and uses reflect.Type.Method which is not implemented.
	@# TestParseAndBytesRoundTrip/P256/Generic: needs Goexit to run defers on wasm.
	@# TestUnmarshalNestingLimit{Slice,Struct}: encoding/asn1 nesting limit added in
	@# https://github.com/golang/go/commit/6a6d115f9a7422b2fa081ba6f567eefb4a099462
	$(TINYGO) test $(TEST_ADDITIONAL_FLAGS) $(TEST_SKIP_FLAG) $(filter-out encoding/xml,$(TEST_PACKAGES_HOST)) $(TEST_PACKAGES_SLOW)
ifeq ($(TEST_ENCODING_XML),true)
	$(TINYGO) test $(TEST_ADDITIONAL_FLAGS) $(TEST_SKIP_FLAG) -stack-size=16MB encoding/xml
endif
	@# io/fs requires os.ReadDir, not yet supported on windows or wasi. It also
	@# requires a large stack-size. Hence, io/fs is only run conditionally.
	@# For more details, see the comments on issue #3143.
ifeq ($(TEST_IOFS),true)
	$(TINYGO) test -stack-size=6MB io/fs
endif
tinygo-test-fast:
	$(TINYGO) test $(TEST_SKIP_FLAG) $(TEST_PACKAGES_HOST)
tinygo-bench:
	$(TINYGO) test -bench . $(TEST_PACKAGES_HOST) $(TEST_PACKAGES_SLOW)
tinygo-bench-fast:
	$(TINYGO) test -bench . $(TEST_PACKAGES_HOST)

# Same thing, except for wasi rather than the current platform.
tinygo-test-wasm:
	$(TINYGO) test -target wasm $(TEST_SKIP_FLAG) $(TEST_PACKAGES_WASM)
tinygo-test-wasi:
	$(TINYGO) test -target wasip1 $(TEST_SKIP_FLAG) $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_SLOW) ./tests/runtime_wasi
tinygo-test-wasip1:
	GOOS=wasip1 GOARCH=wasm $(TINYGO) test $(TEST_SKIP_FLAG) $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_SLOW) ./tests/runtime_wasi
tinygo-test-wasip1-fast:
	$(TINYGO) test -target=wasip1 $(TEST_SKIP_FLAG) $(TEST_PACKAGES_FAST_WASI) ./tests/runtime_wasi

tinygo-test-wasip2-slow:
	$(TINYGO) test -target=wasip2 $(TEST_SKIP_FLAG) $(TEST_PACKAGES_SLOW)
tinygo-test-wasip2-fast:
	$(TINYGO) test -target=wasip2 $(TEST_SKIP_FLAG) $(TEST_PACKAGES_FAST_WASIP2) ./tests/runtime_wasi

tinygo-test-wasip2-sum-slow:
	TINYGO=$(TINYGO) \
	TARGET=wasip2 \
	TESTOPTS="-x -work" \
	PACKAGES="$(TEST_PACKAGES_SLOW)" \
	gotestsum --raw-command -- ./tools/tgtestjson.sh
tinygo-test-wasip2-sum-fast:
	TINYGO=$(TINYGO) \
	TARGET=wasip2 \
	TESTOPTS="-x -work" \
	PACKAGES="$(TEST_PACKAGES_FAST)" \
	gotestsum --raw-command -- ./tools/tgtestjson.sh
tinygo-bench-wasip1:
	$(TINYGO) test -target wasip1 -bench . $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_SLOW)
tinygo-bench-wasip1-fast:
	$(TINYGO) test -target wasip1 -bench . $(TEST_PACKAGES_FAST)

tinygo-bench-wasip2:
	$(TINYGO) test -target wasip2 -bench . $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_SLOW)
tinygo-bench-wasip2-fast:
	$(TINYGO) test -target wasip2 -bench . $(TEST_PACKAGES_FAST)

# Run tests on riscv-qemu since that one provides a large amount of memory.
tinygo-test-baremetal:
	$(TINYGO) test -target riscv-qemu $(TEST_SKIP_FLAG) $(TEST_PACKAGES_BAREMETAL)

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
