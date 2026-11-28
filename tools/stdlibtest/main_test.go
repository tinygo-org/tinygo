package main

import (
	"slices"
	"strings"
	"testing"
)

func TestPackagePlan(t *testing.T) {
	var rules ruleSet
	rules.skipTest("math/rand", `^TestOne\d+$`)
	rules.skipTest("math/rand", `^TestTwo$/^Sub$`)
	rules.stackSizeMB("math/rand", 20)
	rules.short("math/rand")
	if rules.err != nil {
		t.Fatal(rules.err)
	}
	plan := rules.planPackage("math/rand")
	want := []string{"-short", "-stack-size=20MB", `-skip=^TestOne\d+$|^TestTwo$/^Sub$`}
	if got := plan.flags(); !slices.Equal(got, want) {
		t.Fatalf("flags = %v, want %v", got, want)
	}
}

func TestSkipTests(t *testing.T) {
	var hosted ruleSet
	hosted.skipTests("reflect", "TestA")
	hosted.skipTests("reflect", "TestB", "TestC")
	if hosted.err != nil {
		t.Fatal(hosted.err)
	}
	wantHosted := []string{`-skip=^TestA$|^(TestB|TestC)$`}
	if got := hosted.planPackage("reflect").flags(); !slices.Equal(got, wantHosted) {
		t.Fatalf("hosted flags = %v, want %v", got, wantHosted)
	}

	bm := ruleSet{baremetal: true}
	bm.skipTests("reflect", "TestA")
	bm.skipTests("reflect", "TestB", "TestC")
	if bm.err != nil {
		t.Fatal(bm.err)
	}
	wantBM := []string{`-skip=TestA|TestB|TestC`}
	if got := bm.planPackage("reflect").flags(); !slices.Equal(got, wantBM) {
		t.Fatalf("baremetal flags = %v, want %v", got, wantBM)
	}

	for _, add := range []func(*ruleSet){
		func(r *ruleSet) { r.skipTests("reflect") },
		func(r *ruleSet) { r.skipTests("reflect", "^TestA$") },
		func(r *ruleSet) { r.skipTests("reflect", "TestA/Sub") },
	} {
		var bad ruleSet
		add(&bad)
		if bad.err == nil {
			t.Errorf("expected invalid test name error")
		}
	}
}

func TestPackageExclusion(t *testing.T) {
	var rules ruleSet
	rules.skipPackage("crypto/internal/fips140/...")
	rules.skipPackage("crypto/subtle")
	for _, tc := range []struct {
		pkg     string
		skipped bool
	}{
		{"crypto/internal/fips140", true},
		{"crypto/internal/fips140/aes/gcm", true},
		{"crypto/internal/fips140deps", false},
		{"crypto/subtle", true},
		{"crypto/subtle/extra", false},
	} {
		plan := rules.planPackage(tc.pkg)
		if plan.skipped != tc.skipped {
			t.Errorf("%s: skipped = %t, want %t", tc.pkg, plan.skipped, tc.skipped)
		}
	}
}

func TestInvalidRules(t *testing.T) {
	for _, add := range []func(*ruleSet){
		func(r *ruleSet) { r.skipTest("math/rand", "[") },
		func(r *ruleSet) { r.stackSizeMB("math/rand", 0) },
		func(r *ruleSet) { r.skipPackage("crypto/.../aes") },
		func(r *ruleSet) { r.includePackage("./tests/...") },
	} {
		var rules ruleSet
		add(&rules)
		if _, err := planSuite(nil, rules, options{}); err == nil || !strings.Contains(err.Error(), "rules.go:") {
			t.Errorf("%v: expected rule validation error, got %v", rules, err)
		}
	}
	var rules ruleSet
	rules.stackSizeMB("math/rand", 1)
	rules.stackSizeMB("math/rand", 2)
	if _, err := planSuite([]string{"math/rand"}, rules, options{}); err == nil || !strings.Contains(err.Error(), "conflicting stack size") {
		t.Fatalf("expected conflicting stack sizes to fail, got %v", err)
	}
}

func TestSuiteSelection(t *testing.T) {
	t.Run("Windows JSON", func(t *testing.T) {
		windows := targetInfo{GOOS: "windows", GOARCH: "amd64"}
		rules := rulesFor(windows, options{mode: modeFull}, "linux")
		if rules.err != nil {
			t.Fatal(rules.err)
		}
		plan := rules.planPackage("encoding/json")
		if plan.skipped || plan.run != `^TestSynctestMarshal$` {
			t.Fatalf("encoding/json: skipped = %t, run = %q; want only TestSynctestMarshal", plan.skipped, plan.run)
		}
	})

	packages := []string{"compress/bzip2", "os"}
	component := targetInfo{GOOS: "linux", GOARCH: "arm", BuildTags: []string{"tinygo.wasm", "tinygo.wasi", "wasip2"}}
	for _, tc := range []struct {
		name string
		opts options
		want []string
	}{
		{"WASIp2 fast", options{mode: modeFast}, []string{"./tests/runtime_wasi", "os"}},
		{"WASIp2 slow", options{mode: modeSlow}, []string{"compress/bzip2"}},
		{"WASIp2 bench", options{mode: modeFast, bench: true}, []string{"os"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plans, err := planSuite(packages, rulesFor(component, tc.opts, "linux"), tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, p := range plans {
				if !p.skipped {
					got = append(got, p.name)
				}
			}
			if !slices.Equal(got, tc.want) {
				t.Fatalf("selection = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestWASIHostExclusion(t *testing.T) {
	target := targetInfo{GOOS: "wasip1", GOARCH: "wasm", BuildTags: []string{"tinygo.wasm", "tinygo.wasi"}}
	opts := options{mode: modeFull}
	for _, host := range []string{"linux", "windows"} {
		rules := rulesFor(target, opts, host)
		if rules.err != nil {
			t.Fatal(rules.err)
		}
		plan := rules.planPackage("html/template")
		want := []string{`^TestMaxExecDepth$`}
		if host == "windows" {
			want = append(want, `^TestEmptyTemplateHTML$`)
		}
		if !slices.Equal(plan.skips, want) {
			t.Errorf("%s host: skips = %v, want %v", host, plan.skips, want)
		}
	}
}

func TestSuiteShards(t *testing.T) {
	var rules ruleSet
	rules.skipPackage("archive/zip")
	rules.includePackage("./tests/runtime_wasi")
	rules.stackSizeMB("fmt", 1)
	rules.skipTest("time", `^TestTimerGC$`)
	packages := []string{"time", "fmt", "bytes", "archive/zip", "errors", "fmt"}
	for _, tc := range []struct {
		shard int
		want  []string
	}{
		{1, []string{"./tests/runtime_wasi", "errors", "time -skip=^TestTimerGC$"}},
		{2, []string{"bytes", "fmt -stack-size=1MB"}},
	} {
		plans, err := planSuite(packages, rules, options{mode: modeFull, shard: tc.shard, shardCount: 2})
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, p := range plans {
			got = append(got, strings.Join(append([]string{p.name}, p.flags()...), " "))
		}
		if !slices.Equal(got, tc.want) {
			t.Fatalf("shard %d: got %v, want %v", tc.shard, got, tc.want)
		}
	}
}
