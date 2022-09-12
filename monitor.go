package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/mattn/go-tty"
	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/compileopts"
	"go.bug.st/serial"
)

// Monitor connects to the given port and reads/writes the serial port.
func Monitor(port string, options *compileopts.Options) error {
	config, err := builder.NewConfig(options)
	if err != nil {
		return err
	}

	wait := 300
	for i := 0; i <= wait; i++ {
		port, err = getDefaultPort(port, config.Target.SerialPort)
		if err != nil {
			if i < wait {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			return err
		}
		break
	}

	wait = 300
	var p serial.Port
	for i := 0; i <= wait; i++ {
		p, err = serial.Open(port, &serial.Mode{})
		if err != nil {
			if i < wait {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			return err
		}
		break
	}
	defer p.Close()

	tty, err := tty.Open()
	if err != nil {
		return err
	}
	defer tty.Close()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)

	go func() {
		<-sig
		tty.Close()
		os.Exit(0)
	}()

	fmt.Printf("Connected to %s. Press Ctrl-C to exit.\n", port)

	errCh := make(chan error, 1)

	go func() {
		buf := make([]byte, 100*1024)
		for {
			n, err := p.Read(buf)
			if err != nil {
				errCh <- fmt.Errorf("read error: %w", err)
				return
			}

			if n == 0 {
				continue
			}
			fmt.Printf("%v", string(buf[:n]))
		}
	}()

	go func() {
		for {
			r, err := tty.ReadRune()
			if err != nil {
				errCh <- err
				return
			}
			if r == 0 {
				continue
			}
			p.Write([]byte(string(r)))
		}
	}()

	return <-errCh
}
