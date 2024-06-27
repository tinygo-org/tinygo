//go:build baremetal || js || wasm_unknown

package runtime

// This file is for non-hosted environments, that don't support command line
// parameters or environment variables. To still be able to run certain tests,
// command line parameters and environment variables can be passed to the binary
// by setting the variables `runtime.osArgs` and `runtime.osEnv`, both of which
// are strings separated by newlines.
//
// The primary use case is `tinygo test`, which takes some parameters (such as
// -test.v).

// This is the default set of arguments, if nothing else has been set.
var args = []string{"/proc/self/exe"}

//go:linkname os_runtime_args os.runtime_args
func os_runtime_args() []string {
	return args
}

var env []string

//go:linkname syscall_runtime_envs syscall.runtime_envs
func syscall_runtime_envs() []string {
	return env
}

var osArgs string
var osEnv string

func init() {
	if osArgs != "" {
		s := osArgs
		start := 0
		for i := 0; i < len(s); i++ {
			if s[i] == 0 {
				args = append(args, s[start:i])
				start = i + 1
			}
		}
		args = append(args, s[start:])
	}

	if osEnv != "" {
		s := osEnv
		start := 0
		for i := 0; i < len(s); i++ {
			if s[i] == 0 {
				env = append(env, s[start:i])
				start = i + 1
			}
		}
		env = append(env, s[start:])
	}
}
