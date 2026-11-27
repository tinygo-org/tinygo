package main

import (
	"flag"
	"io"
	"os"
	"testing"
)

func TestParent(t *testing.T) {
	t.Run("parallel", func(t *testing.T) {
		t.Parallel()
	})
}

type matchString func(pattern, name string) (bool, error)

func (f matchString) MatchString(pattern, name string) (bool, error) {
	return f(pattern, name)
}

func (f matchString) StartCPUProfile(w io.Writer) error           { return nil }
func (f matchString) StopCPUProfile()                             {}
func (f matchString) WriteProfileTo(string, io.Writer, int) error { return nil }

func main() {
	testing.Init()
	flag.Set("test.v", "true")
	flag.Set("test.parallel", "1")
	m := testing.MainStart(
		matchString(func(string, string) (bool, error) {
			return true, nil
		}),
		[]testing.InternalTest{{Name: "TestParent", F: TestParent}},
		nil,
		nil,
		nil,
	)
	os.Exit(m.Run())
}
