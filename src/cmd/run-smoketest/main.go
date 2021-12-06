package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/sync/errgroup"
)

func main() {
	threads := runtime.NumCPU()
	if j, ok := os.LookupEnv(`RUN_SMOKETEST_JOBS`); ok {
		n, err := strconv.ParseInt(j, 0, 0)
		if err == nil {
			threads = int(n)
		}
	}

	flag.IntVar(&threads, "threads", threads, "threads of make smoketest")
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Printf("usage: %s make smoketest\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	commands, err := getMakeCommands(flag.Args())
	if err != nil {
		log.Fatalf(err.Error())
	}

	err = run(commands, threads)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func getMakeCommands(args []string) ([]string, error) {
	var buf bytes.Buffer
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Args = append(cmd.Args, "-n")
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	commands := []string{}
	scanner := bufio.NewScanner(&buf)

	for scanner.Scan() {
		line := scanner.Text()

		fields := strings.Fields(line)
		if filepath.Base(fields[0]) == "tinygo" && (fields[1] == "build" || fields[1] == "version") {
			commands = append(commands, line)
		} else if strings.HasPrefix(line, "#") {
			commands = append(commands, line)
		}
	}

	return commands, nil
}

func run(commands []string, threads int) error {
	ch := make(chan string, 1000)
	o := NewOchan(ch, 100)

	go func() {
		limit := make(chan struct{}, threads)
		var eg errgroup.Group

		for _, command := range commands {
			command := command
			select {
			case limit <- struct{}{}:
				och := o.GetCh()
				eg.Go(func() error {
					var err error

					if runtime.GOOS == `windows` {
						command = convertToWindowsPath(command)
					}

					fields := strings.Fields(command)
					if filepath.Base(fields[0]) == "tinygo" && fields[1] == "build" {
						err = buildAndmd5sum(command, och)
					} else if strings.HasPrefix(command, "#") {
						och <- fmt.Sprintf("%s\n", command)
					} else {
						err = runCommand(command, och)
					}
					close(och)
					<-limit
					return err
				})
			}
		}

		if err := eg.Wait(); err != nil {
			log.Fatal(err)
		}

		o.Wait()
		close(ch)
	}()

	for s := range ch {
		fmt.Print(s)
	}

	return nil
}

func runCommand(command string, ch chan string) error {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "%s\n", command)

	args := strings.Fields(command)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	if err != nil {
		ch <- buf.String()
		return err
	}

	ch <- buf.String()
	return nil
}

func buildAndmd5sum(command string, ch chan string) error {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "%s\n", command)

	args := strings.Fields(command)

	output := ""
	tmpOutput := ""
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	for i := range args {
		if args[i] == "-o" {
			output = args[i+1]
			ext := filepath.Ext(output)
			tmpOutput = filepath.Join(tmpdir, fmt.Sprintf("%s.%s", filepath.Base(output), ext))
			args[i+1] = tmpOutput
		} else if strings.HasPrefix(args[i], "-o=") {
			output = args[i][3:]
			ext := filepath.Ext(output)
			tmpOutput = filepath.Join(tmpdir, fmt.Sprintf("%s.%s", filepath.Base(output), ext))
			args[i] = fmt.Sprintf("-o=%s", tmpOutput)
		}
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err = cmd.Run()
	if err != nil {
		ch <- buf.String()
		return err
	}

	md5str, err := calcMD5(tmpOutput)
	if err != nil {
		ch <- buf.String()
		return err
	}
	fmt.Fprintf(&buf, "%s  %s\n", md5str, output)

	ch <- buf.String()
	return nil
}

func calcMD5(f string) (string, error) {
	r, err := os.Open(f)
	if err != nil {
		return "", err
	}
	defer r.Close()

	h := md5.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func convertToWindowsPath(command string) string {
	if !strings.HasPrefix(command, `/`) {
		return command
	}

	command = fmt.Sprintf("%c:%s", command[1], command[2:])
	return command
}
