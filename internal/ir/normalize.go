// Package ir compares LLVM IR text in tests across LLVM versions.
package ir

import (
	"regexp"
	"strings"

	"github.com/tinygo-org/tinygo/internal/diff"
)

// Normalize canonicalizes LLVM-version-specific IR spellings for comparison
// and when regenerating golden files.
func Normalize(s string) string {
	// Golden files are written using the pre-LLVM21 'nocapture' spelling,
	// which LLVM printed before any co-occurring attribute such as
	// 'readonly' (e.g. "ptr nocapture readonly"). LLVM 21+ prints the
	// equivalent 'captures(none)' instead, and after such attributes (e.g.
	// "ptr readonly captures(none)"). Normalize both name and position back
	// to the old spelling.
	s = normalizeCapturesAttr(s)

	// LLVM 21+ also added an explicit 'nocreateundeforpoison' attribute to
	// certain intrinsic declarations (e.g. llvm.umin) that were implicitly
	// assumed not to create undef/poison before. It's unrelated to the
	// behavior under test, so ignore it for comparison.
	s = strings.ReplaceAll(s, "nocreateundeforpoison ", "")

	// LLVM 22 dropped the (redundant) i64 size argument from
	// llvm.lifetime.start/end. Normalize away that argument so golden files
	// written against the two-argument form still match.
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

// capturesNoneAttrRe matches a co-occurring attribute directly followed by
// 'captures(none)', which is how LLVM 21+ orders these two attributes when
// printing IR (the pre-LLVM21 'nocapture' attribute printed the other way
// around).
var capturesNoneAttrRe = regexp.MustCompile(`\b(readonly|readnone|writeonly|nonnull)\s+captures\(none\)`)

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

// nofreeNocaptureRe matches LLVM 23's added 'nofree' on a parameter that's
// already 'nocapture'.
var nofreeNocaptureRe = regexp.MustCompile(`\bnofree\s+nocapture\b`)

// normalizeCapturesAttr rewrites LLVM 21+'s 'captures(none)' attribute back
// to the pre-LLVM21 'nocapture' spelling, and drops LLVM 23's added
// 'nosync' attribute and '!guid' metadata, so golden files keep matching
// across versions.
func normalizeCapturesAttr(s string) string {
	s = capturesNoneAttrRe.ReplaceAllString(s, "nocapture $1")
	s = strings.ReplaceAll(s, "captures(none)", "nocapture")
	s = nosyncAttrRe.ReplaceAllString(s, "")
	s = guidAttachmentRe.ReplaceAllString(s, "")
	s = guidMetadataRe.ReplaceAllString(s, "")
	s = targetMemAttrRe.ReplaceAllString(s, "")
	s = nofreeNocaptureRe.ReplaceAllString(s, "nocapture")
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
