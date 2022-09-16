package main

// This file runs some checks on the machine package, to make sure it matches
// the documentation and that it is internally consistent between targets.

import (
	"fmt"
	"go/ast"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/loader"
	"golang.org/x/tools/go/packages"
)

// Constants that are documented and always have a particular type.
var knownConsts = map[string]string{
	// Chip name, something like "nrf51" or "STM32F429".
	"Device": "untyped string",

	// Generic constants.
	"KHz": "untyped int",
	"MHz": "untyped int",
	"GHz": "untyped int",

	// Pin mode types for machine.PinConfig.
	"PinInput":         "machine.PinMode",
	"PinOutput":        "machine.PinMode",
	"PinInputPullup":   "machine.PinMode",
	"PinInputPulldown": "machine.PinMode",

	// Pin change events for pin interrupts.
	"PinRising":  "machine.PinChange",
	"PinFalling": "machine.PinChange",
	"PinToggle":  "machine.PinChange",

	// UART parity
	"ParityNone": "machine.UARTParity",
	"ParityEven": "machine.UARTParity",
	"ParityOdd":  "machine.UARTParity",

	// SPI modes
	"Mode0": "untyped int",
	"Mode1": "untyped int",
	"Mode2": "untyped int",
	"Mode3": "untyped int",

	// Board specific constants.
	"HasLowFrequencyCrystal": "untyped bool", // for nrf boards
}

func main() {
	targets := os.Args[1:]

	// Start worker goroutines to run checkTarget.
	results := make(map[string]chan []error)
	for _, target := range targets {
		results[target] = make(chan []error, 1)
	}
	targetsChan := make(chan string)
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for target := range targetsChan {
				errs := checkTarget(target)
				if len(errs) != 0 {
					results[target] <- errs
				}
				close(results[target])
			}
		}()
	}

	// Check all targets.
	go func() {
		for _, target := range targets {
			targetsChan <- target
		}
	}()

	// Read the result of each check.
	failedTargets := 0
	for _, target := range targets {
		errs := <-results[target]
		if len(errs) != 0 {
			failedTargets++
			fmt.Printf("found errors for %s:\n", target)
			for _, err := range errs {
				fmt.Println(err)
			}
			fmt.Println()
		}
	}

	if failedTargets != 0 {
		fmt.Printf("Failed checks for %d out of %d targets.\n", failedTargets, len(targets))
		os.Exit(1)
	}
}

// Check whether the machine package of the given targets is valid.
func checkTarget(target string) []error {
	// Load target configuration.
	options := &compileopts.Options{
		Target: target,
	}
	spec, err := compileopts.LoadTarget(options)
	if err != nil {
		return []error{err}
	}
	config := &compileopts.Config{Options: options, Target: spec}

	// Load the package.
	goroot, err := loader.GetCachedGoroot(config)
	if err != nil {
		return []error{err}
	}
	cfg := &packages.Config{
		Env: append(os.Environ(),
			"GOOS="+config.GOOS(),
			"GOARCH="+config.GOARCH(),
			"GOROOT="+goroot,
		),
		BuildFlags: []string{
			"-tags=" + strings.Join(spec.BuildTags, " "),
		},
		Mode: packages.NeedSyntax | packages.NeedTypes,
	}
	pkgs, err := packages.Load(cfg, "machine")
	if err != nil {
		return []error{err}
	}
	pkg := pkgs[0]
	if len(pkg.Errors) != 0 {
		var errors []error
		for _, err := range pkg.Errors {
			errors = append(errors, err)
		}
		return errors
	}

	// Check the package!
	checker := Checker{
		goroot: goroot,
		pkg:    pkg,
	}
	for _, file := range pkg.Syntax {
		if checker.getPosition(file.Package).Filename == "i2s.go" {
			// TODO: check I2S.
			continue
		}
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				if isDeprecated(decl.Doc.Text()) {
					// Don't check deprecated declarations. They will be removed
					// with TinyGo 1.0.
					continue
				}
				switch decl.Tok {
				case token.CONST:
					for _, spec := range decl.Specs {
						spec := spec.(*ast.ValueSpec)
						for _, name := range spec.Names {
							if !ast.IsExported(name.Name) {
								continue
							}
							checker.checkConst(spec, name)
						}
					}
				}
			}
		}
	}

	return checker.errors
}

// Return whether the comment is for a deprecated declaration.
// The convention is explained here:
// https://github.com/golang/go/wiki/Deprecated
func isDeprecated(comment string) bool {
	for _, line := range strings.Split(comment, "\n") {
		if strings.HasPrefix(line, "Deprecated: ") {
			return true
		}
	}
	return false
}

type Checker struct {
	goroot string
	pkg    *packages.Package
	errors []error
}

// Check whether a constant declaration follows the documentation.
func (c *Checker) checkConst(spec *ast.ValueSpec, name *ast.Ident) {
	t := c.pkg.Types.Scope().Lookup(name.Name).Type()
	if knownConsts[name.Name] == t.String() {
		return
	}
	switch {
	case t.String() == "machine.Pin":
		// Pins have all kinds of names. They might need some checking, but
		// right now all pin names are fine.
	default:
		c.errors = append(c.errors, scanner.Error{
			Pos: c.getPosition(name.NamePos),
			Msg: fmt.Sprintf("unknown name: %-20s (type %s)", name.Name, t.String()),
		})
	}
}

func (c *Checker) getPosition(pos token.Pos) token.Position {
	position := c.pkg.Fset.Position(pos)
	if newpath, err := filepath.Rel(filepath.Join(c.goroot, "src/machine"), position.Filename); err == nil {
		position.Filename = newpath
	}
	return position
}
