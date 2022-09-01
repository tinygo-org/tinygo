// Stubs for the runtime/trace package
package trace

import (
	"errors"
	"io"
)

func Start(w io.Writer) error {
	return errors.New("not implemented")
}

func Stop() {}
