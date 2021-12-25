package loader

import (
	"os"
	"os/exec"
	"strings"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

// List returns a ready-to-run *exec.Cmd for running the `go list` command with
// the configuration used for TinyGo.
func List(config *compileopts.Config, extraArgs, pkgs []string) (*exec.Cmd, error) {
	goroot, err := GetCachedGoroot(config)
	if err != nil {
		return nil, err
	}
	args := append([]string{"list"}, extraArgs...)
	if len(config.BuildTags()) != 0 {
		args = append(args, "-tags", strings.Join(config.BuildTags(), " "))
	}
	args = append(args, pkgs...)
	cgoEnabled := "0"
	if config.CgoEnabled() {
		cgoEnabled = "1"
	}
	cmd := exec.Command(goenv.Get("GO"), args...)
	cmd.Env = append(os.Environ(), "GOROOT="+goroot, "GOOS="+config.GOOS(), "GOARCH="+config.GOARCH(), "CGO_ENABLED="+cgoEnabled)
	return cmd, nil
}
