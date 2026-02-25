package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// targetSpec is the minimal subset of TinyGo's target JSON we need.
type targetSpec struct {
	Inherits     []string `json:"inherits,omitempty"`
	BuildTags    []string `json:"build-tags,omitempty"`
	GOOS         string   `json:"goos,omitempty"`
	GOARCH       string   `json:"goarch,omitempty"`
	GC           string   `json:"gc,omitempty"`
	Scheduler    string   `json:"scheduler,omitempty"`
	Serial       string   `json:"serial,omitempty"`
	FlashMethod  string   `json:"flash-method,omitempty"`
	FlashCommand string   `json:"flash-command,omitempty"`
	Emulator     string   `json:"emulator,omitempty"`
}

// Target is a resolved target with inheritance applied.
type Target struct {
	Name      string
	BuildTags []string
	GOOS      string
	GOARCH    string
	GC        string
	Scheduler string
	Serial    string
}

// override merges src properties into dst following TinyGo's semantics:
// strings override if non-empty, slices append with dedup.
func (dst *targetSpec) override(src *targetSpec) {
	if src.GOOS != "" {
		dst.GOOS = src.GOOS
	}
	if src.GOARCH != "" {
		dst.GOARCH = src.GOARCH
	}
	if src.GC != "" {
		dst.GC = src.GC
	}
	if src.Scheduler != "" {
		dst.Scheduler = src.Scheduler
	}
	if src.Serial != "" {
		dst.Serial = src.Serial
	}
	if src.FlashMethod != "" {
		dst.FlashMethod = src.FlashMethod
	}
	if src.FlashCommand != "" {
		dst.FlashCommand = src.FlashCommand
	}
	if src.Emulator != "" {
		dst.Emulator = src.Emulator
	}
	dst.BuildTags = appendUnique(dst.BuildTags, src.BuildTags...)
}

func appendUnique(dst []string, src ...string) []string {
	seen := make(map[string]bool, len(dst))
	for _, s := range dst {
		seen[s] = true
	}
	for _, s := range src {
		if !seen[s] {
			dst = append(dst, s)
			seen[s] = true
		}
	}
	return dst
}

func loadRawTargets(dir string) (map[string]*targetSpec, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	specs := make(map[string]*targetSpec)
	for _, e := range entries {
		if !e.Type().IsRegular() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".json")
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var spec targetSpec
		if err := json.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", e.Name(), err)
		}
		specs[name] = &spec
	}
	return specs, nil
}

func resolveSpec(name string, raw map[string]*targetSpec, cache map[string]*targetSpec, resolving map[string]bool) (*targetSpec, error) {
	if cached, ok := cache[name]; ok {
		return cached, nil
	}
	if resolving[name] {
		return nil, fmt.Errorf("circular inheritance: %s", name)
	}
	spec, ok := raw[name]
	if !ok {
		return nil, fmt.Errorf("unknown target: %s", name)
	}

	resolving[name] = true
	defer delete(resolving, name)

	result := &targetSpec{}
	for _, parent := range spec.Inherits {
		resolved, err := resolveSpec(parent, raw, cache, resolving)
		if err != nil {
			return nil, fmt.Errorf("resolving %s parent %s: %w", name, parent, err)
		}
		result.override(resolved)
	}
	result.override(spec)

	cache[name] = result
	return result, nil
}

// LoadTargets reads all target JSONs from dir, resolves inheritance,
// and returns the resolved targets. If includeBase is false, targets
// without flash/emulator configuration are excluded (base/parent targets).
func LoadTargets(dir string, includeBase bool) ([]Target, error) {
	raw, err := loadRawTargets(dir)
	if err != nil {
		return nil, err
	}

	cache := make(map[string]*targetSpec)
	resolving := make(map[string]bool)

	var targets []Target
	for name := range raw {
		resolved, err := resolveSpec(name, raw, cache, resolving)
		if err != nil {
			return nil, err
		}
		if !includeBase && resolved.FlashMethod == "" && resolved.FlashCommand == "" && resolved.Emulator == "" {
			continue
		}
		targets = append(targets, Target{
			Name:      name,
			BuildTags: resolved.BuildTags,
			GOOS:      resolved.GOOS,
			GOARCH:    resolved.GOARCH,
			GC:        resolved.GC,
			Scheduler: resolved.Scheduler,
			Serial:    resolved.Serial,
		})
	}

	sort.Slice(targets, func(i, j int) bool { return targets[i].Name < targets[j].Name })
	return targets, nil
}
