package builder

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/tinygo-org/tinygo/compileopts"
)

// https://infocenter.nordicsemi.com/index.jsp?topic=%2Fug_nrfutil%2FUG%2Fnrfutil%2Fnrfutil_intro.html

func makeDFUFirmwareImage(options *compileopts.Options, infile, outfile string) error {
	cmdLine := []string{"nrfutil", "pkg", "generate", "--hw-version", "52", "--sd-req", "0x0", "--debug-mode", "--application", infile, outfile}

	if options.PrintCommands != nil {
		options.PrintCommands(cmdLine[0], cmdLine[1:]...)
	}

	cmd := exec.Command(cmdLine[0], cmdLine[1:]...)
	cmd.Stdout = io.Discard
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not run nrfutil pkg generate: %w", err)
	}
	return nil
}
