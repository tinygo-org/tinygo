# Test suites: compiler tests, standard library tests, benchmarks, and the corpus.

.PHONY: test wasmtest golden-check

test: check-nodejs-version
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test $(GOTESTFLAGS) -timeout=1h -buildmode exe -tags "byollvm llvm22 osusergo" $(GOTESTPKGS)

golden-check: ## Fail if regenerating the golden IR files changes them
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" $(GO) test -buildmode exe -tags "byollvm llvm22 osusergo" ./compiler ./transform -update
	git diff --exit-code -- compiler/testdata transform/testdata

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
	crypto/des \
	crypto/ecdsa \
	crypto/elliptic \
	crypto/hmac \
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
	encoding/binary \
	encoding/csv \
	encoding/hex \
	encoding/pem \
	expvar \
	go/ast \
	go/format \
	go/scanner \
	go/token \
	go/version \
	hash \
	hash/adler32 \
	hash/crc32 \
	hash/crc64 \
	hash/fnv \
	html \
	html/template \
	image \
	internal/itoa \
	internal/profile \
	internal/reflectlite \
	maps \
	math \
	math/cmplx \
	net/http/internal/ascii \
	net/mail \
	net/url \
	os \
	path \
	reflect \
	regexp/syntax \
	runtime \
	strconv \
	sync \
	testing \
	testing/iotest \
	text/scanner \
	text/tabwriter \
	text/template \
	text/template/parse \
	unicode \
	unicode/utf16 \
	unicode/utf8 \
	unique \
	$(nil)

# archive/zip requires os.ReadAt, which is not yet supported on windows
# bytes requires mmap
# compress/flate appears to hang on wasi
# debug/plan9obj requires os.ReadAt, which is not yet supported on windows
# encoding/xml takes a minute on linux and gives a stack overflow on wasi
# io/ioutil requires os.ReadDir, which is not yet supported on wasi
# mime fails on wasi: bufio.Scanner reports an impossible read count
# mime/multipart: needs wasip1 syscall.FDFLAG_NONBLOCK
# mime/quotedprintable requires syscall.Faccessat
# net/mail: needs wasip1  syscall.FDFLAG_NONBLOCK
# net/ntextproto: needs wasip1 syscall.FDFLAG_NONBLOCK
# testing/fstest requires os.ReadDir, which is not yet supported on wasi

# Additional standard library packages that pass tests on individual platforms
TEST_PACKAGES_LINUX := \
	archive/tar \
	archive/zip \
	bytes \
	compress/flate \
	compress/gzip \
	context \
	crypto/aes \
	crypto/cipher \
	crypto/ecdh \
	crypto/ed25519 \
	crypto/fips140 \
	crypto/hkdf \
	crypto/hpke \
	crypto/mlkem \
	crypto/mldsa \
	crypto/pbkdf2 \
	crypto/rc4 \
	crypto/rsa \
	crypto/sha3 \
	database/sql \
	debug/buildinfo \
	debug/dwarf \
	debug/elf \
	debug/gosym \
	debug/pe \
	debug/plan9obj \
	encoding/gob \
	encoding/json \
	encoding/json/jsontext \
	encoding/json/v2 \
	encoding/xml \
	flag \
	fmt \
	go/build/constraint \
	go/constant \
	go/doc \
	go/doc/comment \
	go/parser \
	go/printer \
	image/color \
	image/draw \
	image/gif \
	image/jpeg \
	image/png \
	io/ioutil \
	iter \
	math/big \
	math/bits \
	math/rand \
	math/rand/v2 \
	mime \
	mime/multipart \
	mime/quotedprintable \
	net \
	net/mail \
	net/netip \
	net/textproto \
	os/user \
	path/filepath \
	regexp \
	slices \
	sort \
	strings \
	testing/cryptotest \
	testing/fstest \
	testing/quick \
	time \
	$(if $(wildcard $(shell $(GO) env GOROOT)/src/uuid),uuid) \
	$(nil)

# The ARM64 DIT tests need Go assembly for internal/runtime/sys.DITEnabled.
ifneq ($(shell $(GO) env GOARCH),arm64)
TEST_PACKAGES_LINUX += crypto/subtle
endif

TEST_PACKAGES_DARWIN := $(TEST_PACKAGES_LINUX)

# The macOS sync/atomic test exits without a diagnostic.
TEST_PACKAGES_LINUX += sync/atomic

# os/user requires t.Skip() support
# time takes more than 25 minutes on Windows.
TEST_PACKAGES_WINDOWS := \
	compress/flate \
	io/ioutil \
	mime \
	testing/fstest \
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
	html/template \
	image \
	os \
	testing \
	text/template \
	$(nil)

# These packages cannot be tested on baremetal.
#
# Some reasons why the tests don't pass on baremetal:
#
#   * No filesystem is available, so packages like compress/zlib can't be tested
#     (just like wasm).
#   * Since Go 1.27 the crypto tests below go through cryptotest.TestHash, which
#     calls cryptotest.BoundarySlices. These targets report GOOS=linux, so they
#     build boundary.go (//go:build linux || darwin) rather than
#     boundary_compat.go, and that needs a working syscall.Mmap/syscall.Mprotect
#     which we don't have. See #5593.
TEST_PACKAGES_BAREMETAL = $(filter-out $(TEST_PACKAGES_NONBAREMETAL), $(TEST_PACKAGES_FAST))
TEST_PACKAGES_NONBAREMETAL = \
	$(TEST_PACKAGES_NONWASM) \
	$(TEST_PACKAGES_NOBOUNDARYSLICES) \
	crypto/des \
	encoding/gob \
	encoding/pem \
	regexp/syntax \
	$(nil)

TEST_PACKAGES_FAST_WASI = $(filter-out $(TEST_PACKAGES_NOWASI), $(TEST_PACKAGES_FAST))
TEST_PACKAGES_NOWASI = \
	$(nil)

# wasip1 reports GOOS=wasip1 and so gets the boundary_compat.go fallback, but
# wasip2 reports GOOS=linux and hits the same BoundarySlices problem as
# baremetal. On wasip2 syscall.Mmap returns ENOSYS and t.Fatalf cannot Goexit,
# so the test falls through and panics with "slice out of range".
TEST_PACKAGES_FAST_WASIP2 = $(filter-out $(TEST_PACKAGES_NOBOUNDARYSLICES), $(TEST_PACKAGES_FAST_WASI))

TEST_PACKAGES_NOBOUNDARYSLICES = \
	crypto/hmac \
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
TEST_PACKAGES_HOST_FULL_ONLY := internal/synctest testing/synctest
ifeq ($(uname),Darwin)
TEST_PACKAGES_HOST := $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_DARWIN)
TEST_PACKAGES_HOST_FULL_ONLY += io/fs
endif
ifeq ($(uname),Linux)
TEST_PACKAGES_HOST := $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_LINUX)
TEST_PACKAGES_HOST_FULL_ONLY += io/fs
endif
ifeq ($(OS),Windows_NT)
TEST_PACKAGES_HOST := $(TEST_PACKAGES_FAST) $(TEST_PACKAGES_WINDOWS)
TEST_PACKAGES_HOST_FULL_ONLY += io/fs
endif

TEST_SKIP_FLAG := -skip='TestAsValidation|TestUnmarshalNestingLimitSlice|TestUnmarshalNestingLimitStruct'
TEST_ADDITIONAL_FLAGS ?=

TEST_OPTIONS_crypto/dsa := -short
TEST_OPTIONS_crypto/ed25519 := -skip='^TestAllocations$$'
TEST_OPTIONS_crypto/rsa := -short -skip='^TestAllocations$$'
TEST_OPTIONS_crypto/sha256 := -skip='^(TestAllocations|TestAllocatonsWithTypeAsserts)$$'
TEST_OPTIONS_crypto/sha3 := -skip='^TestAllocations$$'
TEST_OPTIONS_crypto/sha512 := -skip='^TestAllocations$$'
TEST_OPTIONS_database/sql := -skip='^TestGrabConnAllocs$$'
TEST_OPTIONS_flag := -skip='^TestDefineAfterSet$$'
TEST_OPTIONS_fmt := -stack-size=1MB -skip='^TestCountMallocs$$'
TEST_OPTIONS_sort := -skip='^TestSearchWrappersDontAlloc$$'
TEST_OPTIONS_strconv := -skip='^(TestCountMallocs|TestAllocationsFromBytes)$$'
TEST_OPTIONS_unicode/utf16 := -skip='^(TestAllocationsDecode)$$'
TEST_OPTIONS_math/big := -skip='^TestNewIntAllocs$$'
TEST_OPTIONS_mime := -skip='^TestLookupMallocs$$'
TEST_OPTIONS_reflect := -skip='^TestMapIterSet$$'
TEST_OPTIONS_bytes := -skip='^TestNewBufferShallow$$'
TEST_OPTIONS_slices := -skip='^(TestGrow|TestInsert|TestConcat)$$'
TEST_OPTIONS_strings := -skip='^(TestBuilderAllocs|TestBuilderGrow|TestBuilderGrowSizeclasses|TestIndexRune|TestReplace)$$'
TEST_OPTIONS_crypto/mldsa := -short -stack-size=1MB -skip='^TestAllocations$$'
TEST_OPTIONS_encoding/gob := -skip='^(TestIgnoreDepthLimit|TestCountDecodeMallocs)$$'
TEST_OPTIONS_html/template := -skip='^TestMaxExecDepth$$'
TEST_OPTIONS_text/template := -skip='^TestMaxExecDepth$$'

tinygo-test tinygo-test-fast: TEST_OPTIONS_time := -skip='^TestTimerGC$$'
ifeq ($(uname),Darwin)
# These tests start 1,000 to 5,000 goroutines and can exceed macOS's thread limit.
# https://go.dev/src/time/sleep_test.go and https://go.dev/src/time/tick_test.go
tinygo-test tinygo-test-fast: TEST_OPTIONS_time := -skip='^(TestTimerGC|TestStopResult|TestResetResult|TestLongAdjustTimers)$$|^TestZeroTimer$$/^impl=(func|cache)$$'
# Process creation is a stub on macOS.
tinygo-test tinygo-test-fast: TEST_OPTIONS_flag := -skip='^(TestDefineAfterSet|TestExitCode)$$'
tinygo-test tinygo-test-fast: TEST_OPTIONS_math/rand := -skip='^TestDefaultRace$$'
endif
tinygo-test tinygo-test-fast: TEST_OPTIONS_go/printer := -stack-size=1MB
tinygo-test tinygo-test-fast: TEST_OPTIONS_encoding/gob := -stack-size=1MB -skip='^TestCountDecodeMallocs$$'
tinygo-test tinygo-test-fast: TEST_OPTIONS_encoding/json := -stack-size=20MB -skip='^TestHTTPDecoding$$'
tinygo-test tinygo-test-fast: TEST_OPTIONS_encoding/json/jsontext := -stack-size=20MB -skip='^TestTokenStringAllocations$$'
tinygo-test tinygo-test-fast: TEST_OPTIONS_encoding/json/v2 := -stack-size=20MB
ifeq ($(OS),Windows_NT)
tinygo-test tinygo-test-fast: TEST_OPTIONS_html/template := -skip='^(TestMaxExecDepth|TestParseZipFS)$$'
# The Windows host volume mounted as the guest root lacks /dev/null.
tinygo-test-wasi tinygo-test-wasip1 tinygo-test-wasip1-fast tinygo-test-wasip2-fast: TEST_OPTIONS_html/template := -skip='^(TestMaxExecDepth|TestEmptyTemplateHTML)$$'
else
tinygo-test tinygo-test-fast: TEST_OPTIONS_html/template := -stack-size=20MB
tinygo-test tinygo-test-fast: TEST_OPTIONS_text/template := -stack-size=20MB
endif
tinygo-test tinygo-test-fast: TEST_OPTIONS_testing/quick := -stack-size=1MB
tinygo-test tinygo-test-fast: TEST_OPTIONS_go/build/constraint = $(HOST_TEST_STACK_SKIP_FLAG) -stack-size=1MB
tinygo-test tinygo-test-fast: TEST_OPTIONS_regexp = $(HOST_TEST_STACK_SKIP_FLAG) -stack-size=1MB
tinygo-test tinygo-test-fast: TEST_OPTIONS_path/filepath = $(HOST_TEST_STACK_SKIP_FLAG) -stack-size=4MB
tinygo-test tinygo-test-fast: TEST_OPTIONS_go/parser = $(HOST_TEST_STACK_SKIP_FLAG) -stack-size=256MB
tinygo-test tinygo-test-fast: TEST_OPTIONS_encoding/xml = $(TEST_SKIP_FLAG) -short -stack-size=16MB
tinygo-test: TEST_OPTIONS_archive/zip = $(TEST_SKIP_FLAG) -short
tinygo-test: TEST_OPTIONS_index/suffixarray = $(TEST_SKIP_FLAG) -short
tinygo-test: TEST_OPTIONS_net/netip := -skip='^TestAddrStringAllocs$$|^TestNoAllocs$$/^(Addr.IsGlobalUnicast|Addr.IsInterfaceLocalMulticast|Addr.IsLinkLocalMulticast|Addr.IsLinkLocalUnicast|Addr.IsPrivate)$$'
tinygo-test: TEST_OPTIONS_io/fs := -stack-size=6MB
ifeq ($(OS),Windows_NT)
# TestWalkDir needs File.Chdir, which is a stub in src/os/file_windows.go.
tinygo-test: TEST_OPTIONS_io/fs := -stack-size=8MB -skip='^TestWalkDir$$'
endif
ifeq ($(uname),Linux)
tinygo-test tinygo-test-fast: TEST_OPTIONS_encoding/asn1 := -stack-size=20MB
tinygo-test tinygo-test-fast: TEST_OPTIONS_errors :=
else
TEST_SYNCTEST_BUBBLE_SKIP := |TestChannelMovedOutOfBubble|TestTimerFromInsideBubble|TestWaitGroupMovedIntoBubble|TestWaitGroupMovedOutOfBubble|TestWaitGroupMovedBetweenBubblesWithNonZeroCount
endif
tinygo-test: TEST_OPTIONS_internal/synctest = -skip='^(TestReflectFuncOf$(TEST_SYNCTEST_BUBBLE_SKIP)$(TEST_SYNCTEST_THREAD_LIMIT_SKIP))$$'
tinygo-test: TEST_OPTIONS_testing/synctest := -skip='^(TestFatal|TestError|TestVerboseError|TestSkip|TestVerboseSkip|TestHelper|TestHTTPTransport100Continue)$$'
tinygo-test-fast: HOST_TEST_STACK_SKIP_FLAG = $(TEST_SKIP_FLAG)

tinygo-test-baremetal: TEST_OPTIONS_strconv := -skip='TestCountMallocs|TestAllocationsFromBytes'
tinygo-test-baremetal: TEST_OPTIONS_unicode/utf16 := -skip='TestAllocationsDecode'
tinygo-test-baremetal: TEST_OPTIONS_reflect := -skip='TestMapIterSet'

# Explicit options replace default flags, even when empty.
# https://www.gnu.org/software/make/manual/html_node/Origin-Function.html
test-packages-with-options = $(foreach package,$(1),$(if $(filter undefined,$(origin TEST_OPTIONS_$(package))),,$(package)))

define run-tinygo-test-package
$(3) $(TINYGO) test $(2) $(TEST_OPTIONS_$(1)) $(1)

endef

define run-tinygo-tests
$(if $(strip $(filter-out $(call test-packages-with-options,$(1)),$(1))),$(4) $(TINYGO) test $(2) $(3) $(filter-out $(call test-packages-with-options,$(1)),$(1)))
$(foreach package,$(call test-packages-with-options,$(1)),$(call run-tinygo-test-package,$(package),$(2),$(4)))
endef

# TestWaitGroupManyBubbles can exhaust the Boehm mark stack on Linux
# or the thread limit on macOS.
ifneq ($(filter Linux Darwin,$(uname)),)
TEST_SYNCTEST_THREAD_LIMIT_SKIP := |TestWaitGroupManyBubbles
endif

# Test known-working standard library packages.
# TODO: parallelize, and only show failing tests (no implied -v flag).
.PHONY: tinygo-test
tinygo-test:
	@# TestUnmarshalNestingLimit{Slice,Struct}: encoding/asn1 nesting limit added in
	@# https://github.com/golang/go/commit/6a6d115f9a7422b2fa081ba6f567eefb4a099462
	$(call run-tinygo-tests,$(TEST_PACKAGES_HOST) $(TEST_PACKAGES_SLOW) $(TEST_PACKAGES_HOST_FULL_ONLY),$(TEST_ADDITIONAL_FLAGS),$(TEST_SKIP_FLAG))
ifeq ($(filter encoding/json,$(TEST_PACKAGES_HOST)),)
	$(TINYGO) test $(TEST_ADDITIONAL_FLAGS) -run='^TestSynctestMarshal$$' encoding/json
endif
tinygo-test-fast:
	$(call run-tinygo-tests,$(TEST_PACKAGES_HOST),,$(TEST_SKIP_FLAG))
tinygo-bench:
	$(TINYGO) test -bench . $(TEST_PACKAGES_HOST) $(TEST_PACKAGES_SLOW)
tinygo-bench-fast:
	$(TINYGO) test -bench . $(TEST_PACKAGES_HOST)

# Same thing, except for wasi rather than the current platform.
tinygo-test-wasm:
	$(call run-tinygo-tests,$(TEST_PACKAGES_WASM),-target wasm,$(TEST_SKIP_FLAG))
tinygo-test-wasi:
	$(call run-tinygo-tests,$(TEST_PACKAGES_FAST) $(TEST_PACKAGES_SLOW) ./tests/runtime_wasi,-target wasip1,$(TEST_SKIP_FLAG))
tinygo-test-wasip1:
	$(call run-tinygo-tests,$(TEST_PACKAGES_FAST) $(TEST_PACKAGES_SLOW) ./tests/runtime_wasi,,$(TEST_SKIP_FLAG),GOOS=wasip1 GOARCH=wasm)
tinygo-test-wasip1-fast:
	$(call run-tinygo-tests,$(TEST_PACKAGES_FAST_WASI) ./tests/runtime_wasi,-target=wasip1,$(TEST_SKIP_FLAG))

tinygo-test-wasip2-slow:
	$(call run-tinygo-tests,$(TEST_PACKAGES_SLOW),-target=wasip2,$(TEST_SKIP_FLAG))
tinygo-test-wasip2-fast:
	$(call run-tinygo-tests,$(TEST_PACKAGES_FAST_WASIP2) ./tests/runtime_wasi,-target=wasip2,$(TEST_SKIP_FLAG))

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
	$(call run-tinygo-tests,$(TEST_PACKAGES_BAREMETAL),-target riscv-qemu,$(TEST_SKIP_FLAG))

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
