package builder

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"tinygo.org/x/go-llvm"
)

// Commands lists command alternatives for various operating systems. These
// commands may have a slightly different name across operating systems and
// distributions or may not even exist in $PATH, in which case absolute paths
// may be used.
var commands = map[string][]string{}

func init() {
	llvmMajor := strings.Split(llvm.Version, ".")[0]
	commands["clang"] = []string{"clang-" + llvmMajor}
	commands["ld.lld"] = []string{"ld.lld-" + llvmMajor, "ld.lld"}
	commands["wasm-ld"] = []string{"wasm-ld-" + llvmMajor, "wasm-ld"}
	// Add the path to a Homebrew-installed LLVM for ease of use (no need to
	// manually set $PATH).
	if runtime.GOOS == "darwin" {
		prefix := "/usr/local/opt/llvm@" + llvmMajor + "/bin/"
		commands["clang"] = append(commands["clang"], prefix+"clang-"+llvmMajor)
		commands["ld.lld"] = append(commands["ld.lld"], prefix+"ld.lld")
		commands["wasm-ld"] = append(commands["wasm-ld"], prefix+"wasm-ld")
	}
	// Add the path for when LLVM was installed with the installer from
	// llvm.org, which by default doesn't add LLVM to the $PATH environment
	// variable.
	if runtime.GOOS == "windows" {
		commands["clang"] = append(commands["clang"], "clang", "C:\\Program Files\\LLVM\\bin\\clang.exe")
		commands["ld.lld"] = append(commands["ld.lld"], "lld", "C:\\Program Files\\LLVM\\bin\\lld.exe")
		commands["wasm-ld"] = append(commands["wasm-ld"], "C:\\Program Files\\LLVM\\bin\\wasm-ld.exe")
	}
	// Add the path to LLVM installed from ports.
	if runtime.GOOS == "freebsd" {
		prefix := "/usr/local/llvm" + llvmMajor + "/bin/"
		commands["clang"] = append(commands["clang"], prefix+"clang-"+llvmMajor)
		commands["ld.lld"] = append(commands["ld.lld"], prefix+"ld.lld")
		commands["wasm-ld"] = append(commands["wasm-ld"], prefix+"wasm-ld")
	}
}

func execCommand(cmdNames []string, args ...string) error {
	for _, cmdName := range cmdNames {
		cmd := exec.Command(cmdName, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			if err, ok := err.(*exec.Error); ok && (err.Err == exec.ErrNotFound || err.Err.Error() == "file does not exist") {
				// this command was not found, try the next
				continue
			}
			return err
		}
		return nil
	}
	return errors.New("none of these commands were found in your $PATH: " + strings.Join(cmdNames, " "))
}
