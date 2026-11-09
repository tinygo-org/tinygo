// Package ir compares LLVM IR text in tests across LLVM versions.
package ir

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/tinygo-org/tinygo/internal/diff"
	"tinygo.org/x/go-llvm"
)

// GoldenLLVMVersion is the LLVM major version that writes golden files.
const GoldenLLVMVersion = 22

// WriteGolden writes s unchanged to the golden file at path. It fails unless
// the LLVM in use is GoldenLLVMVersion, so golden files hold only its output.
func WriteGolden(path, s string) error {
	if llvm.VersionMajor != GoldenLLVMVersion {
		return fmt.Errorf("golden files must be written with LLVM %d, not LLVM %s", GoldenLLVMVersion, llvm.Version)
	}
	return os.WriteFile(path, []byte(s), 0666)
}

// Normalize canonicalizes LLVM-version-specific IR spellings, so that output
// from any supported LLVM version can be compared with the golden files.
func Normalize(s string) string {
	// LLVM 21+ prints 'captures(none)' after co-occurring attributes such as
	// 'readonly' (e.g. "ptr readonly captures(none)"), where older LLVM
	// printed 'nocapture' first (e.g. "ptr nocapture readonly"). Normalize
	// both sides to the new spelling so the comparison works on either.
	s = normalizeCapturesAttr(s)

	// LLVM 21+ also added an explicit 'nocreateundeforpoison' attribute to
	// certain intrinsic declarations (e.g. llvm.umin) that were implicitly
	// assumed not to create undef/poison before. It's unrelated to the
	// behavior under test, so ignore it for comparison.
	s = strings.ReplaceAll(s, "nocreateundeforpoison ", "")

	// LLVM 22 dropped the (redundant) i64 size argument from
	// llvm.lifetime.start/end. Normalize away that argument so output from
	// older LLVM matches the golden files.
	s = lifetimeSizeArgRe.ReplaceAllString(s, "$1")

	return s
}

// Diff compares two LLVM IR strings, ignoring irrelevant lines (comments,
// empty lines, etc.) and normalizing LLVM-version-specific spellings via
// Normalize. It returns "" when they are equal, otherwise a unified diff.
func Diff(expected, actual string) string {
	exp := strings.Join(filterIrrelevantLines(strings.Split(Normalize(expected), "\n")), "\n") + "\n"
	act := strings.Join(filterIrrelevantLines(strings.Split(Normalize(actual), "\n")), "\n") + "\n"
	return string(diff.Diff("expected", []byte(exp), "actual", []byte(act)))
}

// nocaptureAttrRe matches the pre-LLVM21 'nocapture' attribute directly
// followed by a co-occurring attribute. LLVM 21+ prints 'captures(none)' after
// such an attribute instead.
var nocaptureAttrRe = regexp.MustCompile(`(\s)nocapture\s+(readonly|readnone|writeonly|nonnull)\b`)

// nocaptureRe matches any other pre-LLVM21 'nocapture' attribute.
var nocaptureRe = regexp.MustCompile(`(\s)nocapture\b`)

// lifetimeSizeArgRe matches the i64 size argument of an
// llvm.lifetime.start/end call or declaration, which LLVM 22 removed.
var lifetimeSizeArgRe = regexp.MustCompile(`(@llvm\.lifetime\.(?:start|end)\.p0\()i64(?: immarg| \d+), `)

// nosyncAttrRe matches the standalone 'nosync' function attribute, which
// LLVM 23 started inferring for llvm.memcpy/llvm.memmove declarations.
var nosyncAttrRe = regexp.MustCompile(`\bnosync\s+`)

// guidAttachmentRe matches the inline '!guid !N' metadata LLVM 23 attaches
// to every GlobalValue once a pass builds a module summary. A global's
// attachment is comma separated, a function's is space separated.
var guidAttachmentRe = regexp.MustCompile(`,?\s*!guid\s+!\d+`)

// guidMetadataRe matches the standalone metadata node backing a '!guid'
// attachment.
var guidMetadataRe = regexp.MustCompile(`(?m)^!\d+ = !\{i64 -?\d+\}\n?`)

// targetMemAttrRe matches the target_mem entries of a memory() attribute.
// Older LLVM numbered them target_mem0/target_mem1; LLVM 23 emits one
// unnumbered target_mem instead.
var targetMemAttrRe = regexp.MustCompile(`(,\s*target_mem\d*: none)+`)

// nofreeCapturesNoneRe matches LLVM 23's added 'nofree' on a parameter that's
// already 'captures(none)'.
var nofreeCapturesNoneRe = regexp.MustCompile(`\bnofree\s+((?:(?:readonly|readnone|writeonly|nonnull)\s+)?captures\(none\))`)

// normalizeCapturesAttr rewrites the pre-LLVM21 'nocapture' attribute to the
// LLVM 21+ 'captures(none)' spelling, and drops LLVM 23's added
// 'nosync' attribute and '!guid' metadata, so golden files keep matching
// across versions.
func normalizeCapturesAttr(s string) string {
	s = nocaptureAttrRe.ReplaceAllString(s, "${1}${2} captures(none)")
	s = nocaptureRe.ReplaceAllString(s, "${1}captures(none)")
	s = nosyncAttrRe.ReplaceAllString(s, "")
	s = guidAttachmentRe.ReplaceAllString(s, "")
	s = guidMetadataRe.ReplaceAllString(s, "")
	s = targetMemAttrRe.ReplaceAllString(s, "")
	s = nofreeCapturesNoneRe.ReplaceAllString(s, "$1")
	return s
}

// filterIrrelevantLines removes lines from the input slice of strings that
// are not relevant in comparing IR. For example, empty lines and comments are
// stripped out.
func filterIrrelevantLines(lines []string) []string {
	var out []string
	for _, line := range lines {
		line = strings.Split(line, ";")[0]  // strip out comments/info
		line = strings.TrimRight(line, " ") // remove trailing spaces from comments
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "source_filename = ") {
			continue
		}
		if strings.HasPrefix(line, "@tinygo.indirect-abi = ") {
			continue
		}
		out = append(out, line)
	}
	return out
}
