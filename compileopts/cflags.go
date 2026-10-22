package compileopts

import "strings"

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
