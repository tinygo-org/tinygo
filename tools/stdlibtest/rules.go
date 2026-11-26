package main

import "slices"

func rulesFor(target targetInfo, opts options, hostOS string) ruleSet {
	var r ruleSet
	hasTag := func(tag string) bool { return slices.Contains(target.BuildTags, tag) }
	baremetal := hasTag("baremetal")
	wasm := hasTag("tinygo.wasm")
	wasi := hasTag("tinygo.wasi")
	browser := wasm && !wasi
	qemu := baremetal && hasTag("virt") && hasTag("qemu")
	wasip1 := target.GOOS == "wasip1"
	wasip2 := hasTag("wasip2")
	hosted := !wasm && !baremetal
	linux := hosted && target.GOOS == "linux"
	darwin := hosted && target.GOOS == "darwin"
	windows := hosted && target.GOOS == "windows"

	// Package exclusions
	// Requires Unix sockets or socket-to-file conversion.
	if !windows && !wasip1 && !browser {
		r.skipPackage("bufio")
	}
	r.skipPackage("os/exec")
	r.skipPackage("syscall")
	// Syslog needs net.OpError.Temporary. SMTP tests need TLS APIs.
	r.skipPackage("log/syslog")
	r.skipPackage("net/smtp")
	// Local listener tests fail or stall.
	r.skipPackage("net/rpc")
	// Tests assume runtime.Compiler is gc or gccgo, or pass it to the go command.
	r.skipPackage("go/build")
	// Go's map implementation needs internal/abi.MapType.
	r.skipPackage("go/importer")
	r.skipPackage("go/internal/gccgoimporter")
	r.skipPackage("go/internal/gcimporter")
	r.skipPackage("go/internal/srcimporter")
	r.skipPackage("go/types")
	r.skipPackage("hash/maphash")
	r.skipPackage("internal/types/errors")
	// OID encoding and self-signed certificate tests fail.
	r.skipPackage("crypto/x509")
	if wasi {
		r.skipTest("io", `^(TestMultiReaderFlatten|TestMultiWriterSingleChainFlatten|TestMultiReaderFreesExhaustedReaders|TestMultiWriter_WriteStringSingleAlloc|TestPipeAllocations)$`)
	} else {
		r.skipPackage("io")
	}
	// The logging tests check the caller's file and line.
	r.skipPackage("log")
	r.skipPackage("log/slog")
	r.skipPackage("testing/slogtest")
	// Signal tests need signal_ignored and runtime_doAllThreadsSyscall.
	r.skipPackage("os/signal")
	// The plugin test programs fail to link libdl.
	r.skipPackage("plugin")
	// Weak pointers stay non-nil after GC and after a finalizer is enqueued.
	r.skipPackage("weak")
	// TinyGo's build tags leave no runnable Linux tests.
	r.skipPackage("embed")
	// BoringCrypto's cache tests reference the missing registerCache symbol.
	r.skipPackage("crypto/internal/boring/...")
	if !linux {
		r.skipPackage("internal/coverage/cfile")
		r.skipPackage("internal/coverage/pods")
		r.skipPackage("internal/coverage/test")
	}
	if !linux && !wasi {
		r.skipPackage("internal/coverage/cformat")
		r.skipPackage("internal/coverage/cmerge")
		r.skipPackage("internal/coverage/slicereader")
		r.skipPackage("internal/coverage/slicewriter")
	}
	// The cgroup tests need stock runtime throw and Linux Syscall6 symbols.
	r.skipPackage("internal/cgrouptest")
	r.skipPackage("internal/runtime/cgroup")
	// The CPU and GC scan tests need Go's x86 CPU assembly.
	r.skipPackage("internal/cpu")
	r.skipPackage("internal/runtime/gc/scan")
	if linux {
		r.skipTest("internal/godebug", `^Test(Metrics|CmdBisect|BisectTestCase)$`)
	} else {
		r.skipPackage("internal/godebug")
	}
	if linux {
		// TestSplicePipePool waits for GC to drain sync.Pool and finalize pipes.
		r.skipTest("internal/poll", `^TestSplicePipePool$`)
	} else {
		r.skipPackage("internal/poll")
	}
	// The stock runtime tests need Go's atomic entry points and internal ABI types.
	r.skipPackage("internal/runtime/atomic")
	r.skipPackage("internal/runtime/maps")
	r.skipPackage("internal/sync")
	// The stock Linux syscall tests need Go's Syscall6 assembly.
	r.skipPackage("internal/runtime/syscall/linux")
	// Contains only cgo benchmarks.
	r.skipPackage("internal/runtime/cgobench")
	// The Wasm interop test uses integer constants too large for TinyGo's int.
	r.skipPackage("syscall/js")
	if !linux {
		r.skipPackage("internal/trace")
		r.skipPackage("internal/trace/testtrace")
	}
	if !linux && !wasi {
		r.skipPackage("internal/trace/internal/tracev1")
		r.skipPackage("internal/trace/tracev2")
	}

	if !linux {
		r.skipPackage("net/rpc/jsonrpc")
	}

	// Inherited exclusions without a confirmed failure.
	r.skipPackage("crypto/internal/fips140/edwards25519")
	r.skipPackage("crypto/internal/fips140cache")
	r.skipPackage("crypto/internal/fips140test")
	if !hosted {
		r.skipPackage("crypto/internal/fips140deps")
		r.skipPackage("os/exec/internal/fdtest")
	}
	if !hosted && !wasip1 {
		r.skipPackage("crypto/internal/sysrand")
	} else if darwin || windows {
		r.skipTest("crypto/internal/sysrand", `^TestReadError$`)
	}
	if !hosted && !wasi {
		r.skipPackage("crypto/internal/fips140/bigmod")
		r.skipPackage("crypto/internal/fips140/rsa")
		r.skipPackage("math/big/internal/asmgen")
	}
	if baremetal {
		r.skipPackage("crypto/internal/cryptotest")
		r.skipPackage("crypto/internal/fips140/edwards25519/field")
		r.skipPackage("crypto/internal/fips140/mldsa")
		r.skipPackage("crypto/internal/fips140only")
	}
	if baremetal {
		r.skipPackage("internal/nettest")
	}
	if !linux {
		r.skipPackage("internal/platform")
	}
	if !linux && !darwin {
		r.skipPackage("internal/syscall/unix")
	}
	if !hosted && !wasi {
		r.skipPackage("internal/testenv")
	}
	if !linux && !darwin && !wasi {
		r.skipPackage("internal/xcoff")
	}
	if (!hosted && !wasi) || (hostOS == "windows" && wasi) {
		r.skipPackage("internal/copyright")
	}
	r.skipPackage("internal/runtime/wasitest")
	r.skipPackage("internal/routebsd")
	r.skipPackage("internal/syscall/windows")
	r.skipPackage("internal/syscall/windows/registry")
	if baremetal {
		r.skipPackage("net/http/internal")
	}
	r.skipPackage("net/internal/cgotest")
	if !linux {
		r.skipPackage("net/internal/socktest")
	}
	if baremetal || browser {
		r.skipTest("embed/internal/embedtest", `TestXGlobal`)
	}
	if baremetal && !qemu {
		r.skipPackage("expvar")
	}

	// Linux and macOS coverage
	if !linux && !darwin {
		if !windows {
			r.skipPackage("encoding/json/jsontext")
			r.skipPackage("encoding/json/v2")
			r.skipPackage("encoding/xml")
		}
		if !windows && !wasip1 {
			r.skipPackage("bytes")
			if !browser {
				r.skipPackage("crypto/cipher")
				r.skipPackage("crypto/sha3")
				r.skipPackage("crypto/subtle")
			}
		}
		if !windows && !wasi {
			r.skipPackage("archive/tar")
			r.skipPackage("compress/gzip")
			r.skipPackage("crypto/ecdh")
			r.skipPackage("crypto/ed25519")
			r.skipPackage("crypto/hpke")
			r.skipPackage("crypto/rsa")
			r.skipPackage("debug/buildinfo")
			r.skipPackage("debug/pe")
			r.skipPackage("encoding/gob")
			r.skipPackage("go/printer")
			r.skipPackage("math/big")
			r.skipPackage("uuid")
			if !browser {
				r.skipPackage("crypto/aes")
				r.skipPackage("crypto/fips140")
				r.skipPackage("crypto/hkdf")
				r.skipPackage("crypto/mlkem")
				r.skipPackage("crypto/pbkdf2")
				r.skipPackage("crypto/rc4")
				r.skipPackage("database/sql")
				r.skipPackage("math/rand")
				r.skipPackage("math/rand/v2")
				r.skipPackage("testing/cryptotest")
			}
			if !browser && !qemu {
				r.skipPackage("go/build/constraint")
				r.skipPackage("testing/quick")
			}
		}
		if !wasi && !browser && !qemu {
			r.skipPackage("os/user")
		}
		if !hosted && !wasi && !browser {
			r.skipPackage("image/gif")
			r.skipPackage("image/jpeg")
			r.skipPackage("image/png")
			r.skipPackage("regexp")
		}
		if !hosted && !wasi && !browser && !qemu {
			r.skipPackage("flag")
			r.skipPackage("fmt")
		}
		if !hosted && !wasi && !browser && !qemu {
			r.skipPackage("slices")
			r.skipPackage("strings")
		}
		r.skipPackage("crypto/mldsa")
		if !wasi {
			r.skipPackage("debug/dwarf")
		}
		if !windows && !wasi {
			r.skipPackage("debug/elf")
		}
		if windows {
			r.skipPackage("debug/pe")
		}
		r.skipPackage("go/parser")
		r.skipPackage("net")
		if !wasm {
			r.skipPackage("net/netip")
			r.skipPackage("net/textproto")
		}
		r.skipPackage("path/filepath")

		if !wasi {
			// Windows File.ReadAt is a stub.
			r.skipPackage("archive/zip")
			r.skipPackage("debug/plan9obj")
		}
		r.skipPackage("time") // Takes more than 25 minutes on Windows.
	}

	if windows {
		r.skipPackage("debug/elf")
		r.skipPackage("mime/multipart") // File.ReadAt is a stub.
	}
	if !hosted && !wasi {
		r.skipPackage("mime/multipart")
	}

	if windows || wasi {
		r.skipTest("flag", `^TestExitCode$`)
	}
	if windows && hasTag("scheduler.tasks") {
		r.skipTest("context", `^TestContextErrDoneRace$`)
	}
	if baremetal {
		r.skipTest("image/draw", `TestPaletted`)
	} else if browser {
		r.skipTest("image/draw", `^TestPaletted$`)
	}
	if wasi {
		r.skipTest("fmt", `^TestScanInts$`)
		r.skipTest("regexp", `^TestOnePassCutoff$`)
		r.skipTest("image/png", `^TestDimensionOverflow$`)
	}

	if wasip2 {
		r.skipTest("archive/tar", `^TestFileInfoHeaderSymlink$`)
		r.skipTest("debug/elf", `^TestNoSectionOverlaps$`)
	}
	if windows || wasip2 {
		r.skipTest("math/rand", `^TestDefaultRace$`)
	}
	if wasi {
		r.skipTest("testing/quick", `^TestMutuallyRecursive$`)
	}
	if browser {
		// The Node runner does not provide the filesystem fixtures.
		r.skipTest("image/gif", `^(TestWriter|TestSubImage|TestEncodeAllGo1Dot4|TestEncodeAllGo1Dot5|TestEncodeAllGo1Dot5GlobalColorModel|TestEncodeWrappedImage)$`)
		r.skipTest("image/jpeg", `^(TestDecodeProgressive|TestDecodeEOF|TestTruncatedSOSDataDoesntPanic|TestIssue56724|TestBadRestartMarker|TestDecodeFlexSubsampling|TestWriter)$`)
		r.skipTest("image/png", `^(TestReader|TestReaderError|TestPalettedDecodeConfig|TestInterlaced|TestPaletted8OutOfRangePixel|TestWriter|TestDimensionOverflow)$`)
		r.skipTest("regexp", `^(TestRE2Exhaustive|TestRE2Search)$`)
	}

	// Hosted coverage
	if !hosted {
		r.skipPackage("encoding/json")
		r.skipPackage("internal/synctest")
		r.skipPackage("testing/synctest")
	}
	if !hosted && !wasi {
		r.skipPackage("io/fs")
		r.skipPackage("io/ioutil")
		r.skipPackage("testing/fstest")
		r.skipPackage("mime")
	}
	if wasip2 {
		r.skipTest("io/ioutil", `^TestReadOnlyWriteFile$`)
		r.skipTest("testing/fstest", `^TestSymlink$`)
	}

	// ARM64 DIT tests need Go assembly for internal/runtime/sys.DITEnabled.
	if target.GOARCH == "arm64" {
		r.skipPackage("crypto/subtle")
	}
	// The atomic suite exits without a diagnostic on macOS.
	if !linux && !windows {
		r.skipPackage("sync/atomic")
	}

	// These suites exercise a filesystem or read fixtures from disk.
	if baremetal || (wasm && !wasi) {
		r.skipPackage("compress/flate")
		r.skipPackage("compress/lzw")
		r.skipPackage("compress/zlib")
		r.skipPackage("crypto/ecdsa")
		r.skipPackage("debug/gosym")
		r.skipPackage("debug/macho")
		r.skipPackage("go/doc")
		r.skipPackage("go/doc/comment")
		r.skipPackage("go/format")
		r.skipPackage("html/template")
		r.skipPackage("image")
		r.skipPackage("internal/diff")
		r.skipPackage("os")
		r.skipPackage("testing")
		r.skipPackage("text/template")
	}
	// Go 1.27's DES boundary tests need mprotect.
	if baremetal {
		r.skipPackage("crypto/des")
		// Both parser test binaries exhaust the QEMU heap.
		r.skipPackage("encoding/pem")
		r.skipPackage("regexp/syntax")
	}

	// Go 1.27 hash boundary tests need mmap and mprotect on targets reporting
	// GOOS=linux, including WASIp2 and QEMU. See tinygo-org/tinygo#5593.
	if baremetal || wasip2 {
		r.skipPackage("crypto/hmac")
		r.skipPackage("crypto/md5")
		r.skipPackage("crypto/sha1")
		r.skipPackage("crypto/sha256")
		r.skipPackage("crypto/sha512")
	}

	// Suite modes and fixtures
	// These suites take over a minute on WASI.
	if opts.mode == modeFast {
		r.skipPackage("compress/bzip2")
		r.skipPackage("crypto/dsa")
		r.skipPackage("index/suffixarray")
	}
	if opts.mode == modeSlow {
		r.onlyPackages("compress/bzip2", "crypto/dsa", "index/suffixarray")
	}
	if opts.mode != modeFull || opts.bench {
		r.skipPackage("internal/synctest")
		r.skipPackage("testing/synctest")
	}
	if opts.bench || (opts.mode != modeFull && !wasi) {
		r.skipPackage("io/fs")
	}
	if !opts.bench && (wasip1 || (wasip2 && opts.mode == modeFast)) {
		r.includePackage("./tests/runtime_wasi")
	}

	// Windows runs only the JSON synctest marshal test, in the full suite.
	if windows {
		if opts.mode == modeFull && !opts.bench {
			r.runTest("encoding/json", `^TestSynctestMarshal$`)
		} else {
			r.skipPackage("encoding/json")
		}
	}

	if (opts.mode == modeFull || wasi) && !opts.bench {
		r.short("archive/zip")
	}
	if !opts.bench {
		if wasi {
			r.short("compress/flate")
		}
		r.short("crypto/dsa")
		r.short("crypto/mldsa")
		r.short("crypto/rsa")
		r.short("encoding/xml")
	}
	if hosted && !opts.bench {
		r.short("index/suffixarray")
	}

	if linux {
		r.stackSizeMB("encoding/asn1", 20)
	}
	r.stackSizeMB("crypto/mldsa", 1)
	r.stackSizeMB("encoding/gob", 1)
	if !windows {
		r.stackSizeMB("encoding/json", 20)
	}
	r.stackSizeMB("encoding/json/jsontext", 20)
	r.stackSizeMB("encoding/json/v2", 20)
	r.stackSizeMB("encoding/xml", 16)
	r.stackSizeMB("fmt", 1)
	r.stackSizeMB("go/build/constraint", 1)
	r.stackSizeMB("go/parser", 256)
	r.stackSizeMB("go/printer", 1)
	if linux || darwin {
		r.stackSizeMB("html/template", 20)
	}
	if windows {
		r.stackSizeMB("io/fs", 8)
	} else {
		r.stackSizeMB("io/fs", 6)
	}
	if wasi {
		// The deep glob test exhausts the Wasm stack.
		r.skipTest("io/fs", `^TestCVE202230630$`)
	}
	r.stackSizeMB("path/filepath", 4)
	r.stackSizeMB("regexp", 1)
	r.stackSizeMB("testing/quick", 1)
	if linux || darwin {
		r.stackSizeMB("text/template", 20)
	}

	// TinyGo exceeds the upstream allocation budgets for these operations.
	r.skipTest("bufio", `^TestReadStringAllocs$`)
	r.skipTest("bytes", `^TestNewBufferShallow$`)
	r.skipTest("crypto/ed25519", `^TestAllocations$`)
	r.skipTest("crypto/rsa", `^TestAllocations$`)
	r.skipTest("crypto/sha256", `^(TestAllocations|TestAllocatonsWithTypeAsserts)$`)
	r.skipTest("crypto/sha3", `^TestAllocations$`)
	r.skipTest("crypto/sha512", `^TestAllocations$`)
	r.skipTest("database/sql", `^TestGrabConnAllocs$`)
	r.skipTest("encoding/gob", `^TestCountDecodeMallocs$`)
	r.skipTest("encoding/json/jsontext", `^TestTokenStringAllocations$`)
	r.skipTest("math/big", `^TestNewIntAllocs$`)
	r.skipTest("mime", `^TestLookupMallocs$`)
	if opts.mode == modeFull || opts.mode == modeFast || wasm {
		r.skipTest("net/netip", `^TestAddrStringAllocs$|^TestNoAllocs$/^(Addr.IsGlobalUnicast|Addr.IsInterfaceLocalMulticast|Addr.IsLinkLocalMulticast|Addr.IsLinkLocalUnicast|Addr.IsPrivate)$|^TestNoAllocs$/^Addr\.Zone$/^6zone$`)
	}
	if !baremetal {
		r.skipTest("reflect", `^TestMapIterSet$`)
	}
	if !baremetal {
		r.skipTest("sort", `^TestSearchWrappersDontAlloc$`)
		r.skipTest("strconv", `^(TestCountMallocs|TestAllocationsFromBytes)$`)
	}
	if !baremetal {
		r.skipTest("unicode/utf16", `^TestAllocationsDecode$`)
	}

	// The same allocation checks on bare metal use substring matches.
	if baremetal {
		r.skipTest("reflect", `TestMapIterSet`)
		r.skipTest("sort", `TestSearchWrappersDontAlloc`)
		r.skipTest("strconv", `TestCountMallocs|TestAllocationsFromBytes`)
		r.skipTest("unicode/utf16", `TestAllocationsDecode`)
	}
	if qemu {
		r.skipTest("fmt", `TestCountMallocs`)
		r.skipTest("slices", `TestGrow|TestInsert|TestConcat`)
		r.skipTest("strings", `TestBuilderAllocs|TestBuilderGrow|TestBuilderGrowSizeclasses|TestIndexRune|TestReplace`)
	} else {
		r.skipTest("fmt", `^TestCountMallocs$`)
		r.skipTest("slices", `^(TestGrow|TestInsert|TestConcat)$`)
		r.skipTest("strings", `^(TestBuilderAllocs|TestBuilderGrow|TestBuilderGrowSizeclasses|TestIndexRune|TestReplace)$`)
	}

	if baremetal {
		// TestLarge also matches TestLargeXXHash.
		r.skipTest("internal/zstd", `TestAlloc|TestLarge|TestFileSamples`)
	} else {
		r.skipTest("internal/zstd", `^TestAlloc$`)
		if !linux {
			r.skipTest("internal/zstd", `^TestLarge$`)
		}
		if browser {
			r.skipTest("internal/zstd", `^(TestFileSamples|TestLargeXXHash)$`)
		}
	}

	// TestPureGoTag calls go/build, which rejects runtime.Compiler == "tinygo".
	// TestDisallowedAssemblyInstructions walks GOROOT/src/crypto on disk.
	if baremetal {
		r.skipTest("crypto", `TestPureGoTag|TestDisallowedAssemblyInstructions`)
	} else if browser || (hostOS == "windows" && wasi) {
		r.skipTest("crypto", `^(TestPureGoTag|TestDisallowedAssemblyInstructions)$`)
	} else {
		r.skipTest("crypto", `^TestPureGoTag$`)
	}
	// The ML-DSA allocation check times out.
	r.skipTest("crypto/mldsa", `^TestAllocations$`)
	// The ASN.1 nesting tests can exhaust fixed stacks. Linux uses a 20 MB stack.
	if !linux {
		r.skipTest("encoding/asn1", `TestUnmarshalNestingLimitSlice|TestUnmarshalNestingLimitStruct`)
	}
	// TestHTTPDecoding cannot reach its local HTTP server.
	if !windows {
		r.skipTest("encoding/json", `^TestHTTPDecoding$`)
	}
	// The warning for late flag definitions needs the caller's file and line.
	if qemu {
		r.skipTest("flag", `TestExitCode|TestDefineAfterSet`)
	} else {
		r.skipTest("flag", `^TestDefineAfterSet$`)
	}
	// Boehm does not reclaim unreachable timers and tickers in this test.
	r.skipTest("time", `^TestTimerGC$`)

	// Both tests start subprocesses. macOS startProcess is a stub.
	if darwin {
		r.skipTest("flag", `^TestExitCode$`)
		r.skipTest("math/rand", `^TestDefaultRace$`)
		// Timer stress tests can exhaust macOS threads.
		r.skipTest("time", `^(TestStopResult|TestResetResult|TestLongAdjustTimers)$|^TestZeroTimer$/^impl=(func|cache)$`)
	}
	// This test can exhaust Boehm's mark stack or Darwin's thread limit.
	if hasTag("gc.boehm") || darwin {
		r.skipTest("internal/synctest", `^TestWaitGroupManyBubbles$`)
	}
	// Concurrent FuncOf and StructOf creation times out.
	r.skipTest("internal/synctest", `^TestReflectFuncOf$`)
	// Bubble ownership checks start subprocesses.
	if !linux {
		r.skipTest("internal/synctest", `^(TestChannelMovedOutOfBubble|TestTimerFromInsideBubble|TestWaitGroupMovedIntoBubble|TestWaitGroupMovedOutOfBubble|TestWaitGroupMovedBetweenBubblesWithNonZeroCount)$`)
	}
	// The output cases need bubble result propagation and caller/helper locations.
	// The HTTP case needs Transport.DialContext and ExpectContinueTimeout.
	r.skipTest("testing/synctest", `^(TestFatal|TestError|TestVerboseError|TestSkip|TestVerboseSkip|TestHelper|TestHTTPTransport100Continue)$`)

	// Template depth checks run with larger stacks on Linux and macOS.
	if !linux && !darwin {
		r.skipTest("html/template", `^TestMaxExecDepth$`)
		r.skipTest("text/template", `^TestMaxExecDepth$`)
	}
	// The ZIP filesystem needs File.ReadAt, a stub on Windows.
	if windows {
		r.skipTest("html/template", `^TestParseZipFS$`)
	}
	// The Windows host volume mounted as the WASI root lacks /dev/null.
	if hostOS == "windows" && (wasip1 || wasip2) {
		r.skipTest("html/template", `^TestEmptyTemplateHTML$`)
	}

	return r
}
