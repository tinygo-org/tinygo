package main

// RewriteSpecs lists the packages TinyGo generates from upstream.
var RewriteSpecs = []RewriteSpec{
	{
		Path: "internal/diff",
		// tinygo fix shows its changes with Diff, but cannot import a GOROOT
		// internal package.
		Dir:   "internal/diff",
		Steps: Extract("Diff"),
	},
	{
		Path: "internal/strconv",
		Steps: Pipeline(
			// optimize is false, so the pow10 tables are not linked in.
			Walk(
				ConstFalse("optimize"),
			),
			Prune(),
		),
	},
	{
		Path: "unicode/utf8",
		Steps: Pipeline(
			// ptrSize is 2 on 16-bit targets such as AVR.
			Walk(
				ReplaceDecl("ptrSize", `const ptrSize = 1 << (^uintptr(0)>>15&1 + ^uintptr(0)>>31&1 + ^uintptr(0)>>63&1)`),
			),
			// word loads 2 bytes when ptrSize is 2.
			Walk(
				InsertFirst("word", `if ptrSize == 2 { return uintptr(s[0]) | uintptr(s[1])<<8 }`),
			),
		),
	},
}
