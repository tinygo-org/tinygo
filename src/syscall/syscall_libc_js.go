//go:build js

package syscall

// wasi-libc keeps its own working directory, which starts at "/". Start in
// the directory the host names in $PWD instead.
func init() {
	if wd, ok := Getenv("PWD"); ok && len(wd) > 0 && wd[0] == '/' {
		Chdir(wd)
	}
}
