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
	commands["lldb"] = []string{"lldb-" + llvmMajor, "lldb"}
	// Add the path to a Homebrew-installed LLVM for ease of use (no need to
	// manually set $PATH).
	if runtime.GOOS == "darwin" {
		prefix := "/usr/local/opt/llvm@" + llvmMajor + "/bin/"
		commands["clang"] = append(commands["clang"], prefix+"clang-"+llvmMajor)
		commands["ld.lld"] = append(commands["ld.lld"], prefix+"ld.lld")
		commands["wasm-ld"] = append(commands["wasm-ld"], prefix+"wasm-ld")
		commands["lldb"] = append(commands["lldb"], prefix+"lldb")
	}
	// Add the path for when LLVM was installed with the installer from
	// llvm.org, which by default doesn't add LLVM to the $PATH environment
	// variable.
	if runtime.GOOS == "windows" {
		commands["clang"] = append(commands["clang"], "clang", "C:\\Program Files\\LLVM\\bin\\clang.exe")
		commands["ld.lld"] = append(commands["ld.lld"], "lld", "C:\\Program Files\\LLVM\\bin\\lld.exe")
		commands["wasm-ld"] = append(commands["wasm-ld"], "C:\\Program Files\\LLVM\\bin\\wasm-ld.exe")
		commands["lldb"] = append(commands["lldb"], "C:\\Program Files\\LLVM\\bin\\lldb.exe")
	}
	// Add the path to LLVM installed from ports.
	if runtime.GOOS == "freebsd" {
		prefix := "/usr/local/llvm" + llvmMajor + "/bin/"
		commands["clang"] = append(commands["clang"], prefix+"clang-"+llvmMajor)
		commands["ld.lld"] = append(commands["ld.lld"], prefix+"ld.lld")
		commands["wasm-ld"] = append(commands["wasm-ld"], prefix+"wasm-ld")
		commands["lldb"] = append(commands["lldb"], prefix+"lldb")
	}
}

// LookupCommand looks up the executable name for a given LLVM tool such as
// clang or wasm-ld. It returns the (relative) command that can be used to
// invoke the tool or an error if it could not be found.
func LookupCommand(name string) (string, error) {
	for _, cmdName := range commands[name] {
		_, err := exec.LookPath(cmdName)
		if err != nil {
			if errors.Unwrap(err) == exec.ErrNotFound {
				continue
			}
			return cmdName, err
		}
		return cmdName, nil
	}
	return "", errors.New("%#v: none of these commands were found in your $PATH: " + strings.Join(commands[name], " "))
}

func execCommand(name string, args ...string) error {
	name, err := LookupCommand(name)
	if err != nil {
		return err
	}
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
