// Command update-binary-sizes applies the diff printed by a failed
// TestBinarySize run to builder/testdata/binary-size.txt.
//
// Usage:
//
//	update-binary-sizes [-f file] [log-file | job-url | run-url]
//
// The input is raw go test output or a GitHub Actions log. It is read from
// stdin if no argument is given. A URL of a failed job or run is fetched
// with gh, or with the GITHUB_TOKEN environment variable if gh is missing.
// For example:
//
//	go run ./tools/update-binary-sizes https://github.com/tinygo-org/tinygo/actions/runs/1/job/2
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// A diff line is eight spaces, a marker, then the row. GitHub logs put a
// timestamp ending in Z before the eight spaces.
var diffLine = regexp.MustCompile(`(?:^|Z )\s{8}([-+])(\S.*)$`)

// A row is target, package, optimization, then four numbers.
var rowRe = regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+\d+\s+\d+\s+\d+\s+\d+\s*$`)

// A job or run URL: repo, run id, optional job id.
var urlRe = regexp.MustCompile(`^https://github\.com/([^/]+/[^/]+)/actions/runs/(\d+)(?:/job/(\d+))?`)

// fetchLog returns the failed log of a job, or of every failed job in a run.
func fetchLog(repo, run, job string) (io.ReadCloser, error) {
	args := []string{"run", "view", run, "--repo", repo, "--log-failed"}
	if job != "" {
		args = []string{"run", "view", "--repo", repo, "--job", job, "--log-failed"}
	}
	if _, err := exec.LookPath("gh"); err == nil {
		out, err := exec.Command("gh", args...).Output()
		if err != nil {
			return nil, fmt.Errorf("gh %s: %w", strings.Join(args, " "), err)
		}
		return io.NopCloser(strings.NewReader(string(out))), nil
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" || job == "" {
		return nil, fmt.Errorf("gh not found; set GITHUB_TOKEN and pass a job URL")
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/actions/jobs/%s/logs", repo, job), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("fetching job log: %s", resp.Status)
	}
	return resp.Body, nil
}

func key(row string) string {
	m := rowRe.FindStringSubmatch(row)
	if m == nil {
		return ""
	}
	return m[1] + " " + m[2] + " " + m[3]
}

func main() {
	file := flag.String("f", "builder/testdata/binary-size.txt", "size file to update")
	flag.Parse()

	var in io.Reader = os.Stdin
	if flag.NArg() > 0 {
		if m := urlRe.FindStringSubmatch(flag.Arg(0)); m != nil {
			log, err := fetchLog(m[1], m[2], m[3])
			if err != nil {
				fatal(err)
			}
			defer log.Close()
			in = log
		} else {
			f, err := os.Open(flag.Arg(0))
			if err != nil {
				fatal(err)
			}
			defer f.Close()
			in = f
		}
	}

	removed := map[string]bool{}
	added := map[string]string{}
	var order []string
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		m := diffLine.FindStringSubmatch(strings.TrimRight(sc.Text(), "\r"))
		if m == nil {
			continue
		}
		k := key(m[2])
		if k == "" {
			continue
		}
		if m[1] == "-" {
			removed[k] = true
		} else {
			if _, ok := added[k]; !ok {
				order = append(order, k)
			}
			added[k] = m[2]
		}
	}
	if err := sc.Err(); err != nil {
		fatal(err)
	}
	if len(removed) == 0 && len(added) == 0 {
		fatal(fmt.Errorf("no size diff found in input"))
	}

	data, err := os.ReadFile(*file)
	if err != nil {
		fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	var out []string
	changed, deleted := 0, 0
	for _, l := range lines {
		k := key(l)
		if k == "" {
			out = append(out, l)
			continue
		}
		if row, ok := added[k]; ok {
			out = append(out, row)
			delete(added, k)
			changed++
		} else if removed[k] {
			deleted++
		} else {
			out = append(out, l)
		}
	}
	appended := 0
	for _, k := range order {
		if row, ok := added[k]; ok {
			out = append(out, row)
			appended++
		}
	}
	if err := os.WriteFile(*file, []byte(strings.Join(out, "\n")+"\n"), 0o666); err != nil {
		fatal(err)
	}
	fmt.Fprintf(os.Stderr, "%s: %d changed, %d removed, %d appended\n", *file, changed, deleted, appended)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "update-binary-sizes:", err)
	os.Exit(1)
}
