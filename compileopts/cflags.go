package compileopts

import (
	"path/filepath"
	"strings"
)

// FilePrefixMap returns a Clang file-prefix map for two directory paths.
func FilePrefixMap(from, to string) string {
	return directoryPrefixMap("-ffile-prefix-map=", from, to)
}

// DebugPrefixMap returns a Clang debug-prefix map for two directory paths.
func DebugPrefixMap(from, to string) string {
	return directoryPrefixMap("-fdebug-prefix-map=", from, to)
}

func directoryPrefixMap(flag, from, to string) string {
	if !strings.HasSuffix(from, string(filepath.Separator)) {
		from += string(filepath.Separator)
	}
	if !strings.HasSuffix(to, "/") {
		to += "/"
	}
	return flag + from + "=" + to
}

// RewriteCFlags rewrites path operands without changing other compiler options.
func RewriteCFlags(flags []string, rewrite func(string) string) []string {
	flags = append([]string(nil), flags...)
	nextIsPath := false
	for i, flag := range flags {
		if nextIsPath {
			flags[i] = rewrite(flag)
			nextIsPath = false
			continue
		}
		switch flag {
		case "-I", "-L", "-F", "-isystem", "-isystem-after", "-iquote", "-idirafter", "-include", "--include", "-imacros", "--imacros", "-include-pch", "-isysroot", "--sysroot", "-resource-dir", "-iframework", "-iframeworkwithsysroot":
			nextIsPath = true
			continue
		}
		for _, prefix := range []string{
			"-I", "-L", "-F",
			"-isystem-after", "-isystem", "-iquote", "-idirafter",
			"-iframeworkwithsysroot", "-iframework",
			"-include", "--include", "-imacros", "--imacros",
			"--sysroot=", "-isysroot=", "-isysroot", "-resource-dir=",
			"-fmodule-map-file=", "-fmodules-cache-path=",
		} {
			if after, ok := strings.CutPrefix(flag, prefix); ok {
				flags[i] = prefix + rewrite(after)
				break
			}
		}
		for _, prefix := range []string{"-fdebug-prefix-map=", "-ffile-prefix-map=", "-fmacro-prefix-map="} {
			if mapping, ok := strings.CutPrefix(flag, prefix); ok {
				if sep := strings.LastIndexByte(mapping, '='); sep >= 0 {
					flags[i] = prefix + rewrite(mapping[:sep]) + mapping[sep:]
				}
				break
			}
		}
	}
	return flags
}
