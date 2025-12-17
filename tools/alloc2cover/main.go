package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: alloc2cover <input-file> <source-root> [module-path]")
		fmt.Fprintln(os.Stderr, "  Reads TinyGo -print-allocs output and converts to cover.out format")
		fmt.Fprintln(os.Stderr, "  source-root: path to the root of the source files")
		fmt.Fprintln(os.Stderr, "  module-path: optional Go module path prefix (e.g., github.com/user/repo)")
		fmt.Fprintln(os.Stderr, "  Heap allocations are marked as uncovered (red in coverage reports)")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	sourceRoot := os.Args[2]
	modulePath := ""
	if len(os.Args) > 3 {
		modulePath = os.Args[3]
	}

	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Pattern to match allocation lines
	// /path/to/file.go:line:column: object allocated on the heap: reason
	pattern := regexp.MustCompile(`^(.+\.go):(\d+):\d+: object allocated on the heap:`)

	type allocation struct {
		absPath string
		modPath string
		line    int
	}

	allocations := make(map[string]allocation)
	fileLineLengths := make(map[string][]int) // cache of line lengths per file

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := pattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		absPath := matches[1]
		lineNum, _ := strconv.Atoi(matches[2])

		// Convert absolute path to module path if provided
		modPath := absPath
		if modulePath != "" {
			idx := strings.Index(absPath, modulePath)
			if idx >= 0 {
				modPath = absPath[idx:]
			}
		}

		alloc := allocation{
			absPath: absPath,
			modPath: modPath,
			line:    lineNum,
		}

		key := fmt.Sprintf("%s:%d", modPath, lineNum)
		allocations[key] = alloc
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Helper function to get line lengths for a file
	getLineLengths := func(absPath, modPath string) []int {
		if lengths, ok := fileLineLengths[absPath]; ok {
			return lengths
		}

		// Try to find the file
		var filePath string
		if _, err := os.Stat(absPath); err == nil {
			filePath = absPath
		} else if sourceRoot != "" {
			// Try relative to source root using module path
			filePath = filepath.Join(sourceRoot, modPath)
		}

		if filePath == "" {
			return nil
		}

		f, err := os.Open(filePath)
		if err != nil {
			return nil
		}
		defer f.Close()

		var lengths []int
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			lengths = append(lengths, len(scanner.Text()))
		}
		fileLineLengths[absPath] = lengths
		return lengths
	}

	// Output in cover.out format
	fmt.Println("mode: set")

	// Sort keys for consistent output
	keys := make([]string, 0, len(allocations))
	for k := range allocations {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		a := allocations[key]

		endCol := 1 // default if we can't read the file

		// Try to get actual line length for end column
		lengths := getLineLengths(a.absPath, a.modPath)
		if lengths != nil && a.line <= len(lengths) {
			endCol = lengths[a.line-1] + 1 // +1 for end of line
		}

		// Format: file:line.column,line.column statements count
		// start at column 1, end at actual line length
		// count = 0 marks as uncovered (red in coverage reports)
		fmt.Printf("%s:%d.1,%d.%d 1 0\n", a.modPath, a.line, a.line, endCol)
	}
}
