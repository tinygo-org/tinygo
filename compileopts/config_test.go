package compileopts

import (
	"fmt"
	"strings"
	"testing"
)

func TestBuildTags(t *testing.T) {
	tests := []struct {
		targetTags []string
		userTags   []string
		result     []string
	}{
		{
			targetTags: []string{},
			userTags:   []string{},
			result: []string{
				"tinygo",
				"math_big_pure_go",
				"gc.conservative",
				"scheduler.none",
				"serial.none",
			},
		},
		{
			targetTags: []string{"bear"},
			userTags:   []string{},
			result: []string{
				"bear",
				"tinygo",
				"math_big_pure_go",
				"gc.conservative",
				"scheduler.none",
				"serial.none",
			},
		},
		{
			targetTags: []string{},
			userTags:   []string{"cat"},
			result: []string{
				"tinygo",
				"math_big_pure_go",
				"gc.conservative",
				"scheduler.none",
				"serial.none",
				"cat",
			},
		},
		{
			targetTags: []string{"bear"},
			userTags:   []string{"cat"},
			result: []string{
				"bear",
				"tinygo",
				"math_big_pure_go",
				"gc.conservative",
				"scheduler.none",
				"serial.none",
				"cat",
			},
		},
		{
			targetTags: []string{"bear", "runtime_memhash_leveldb"},
			userTags:   []string{"cat"},
			result: []string{
				"bear",
				"runtime_memhash_leveldb",
				"tinygo",
				"math_big_pure_go",
				"gc.conservative",
				"scheduler.none",
				"serial.none",
				"cat",
			},
		},
		{
			targetTags: []string{"bear", "runtime_memhash_leveldb"},
			userTags:   []string{"cat", "runtime_memhash_leveldb"},
			result: []string{
				"bear",
				"tinygo",
				"math_big_pure_go",
				"gc.conservative",
				"scheduler.none",
				"serial.none",
				"cat",
				"runtime_memhash_leveldb",
			},
		},
		{
			targetTags: []string{"bear", "runtime_memhash_leveldb"},
			userTags:   []string{"cat", "runtime_memhash_sip"},
			result: []string{
				"bear",
				"tinygo",
				"math_big_pure_go",
				"gc.conservative",
				"scheduler.none",
				"serial.none",
				"cat",
				"runtime_memhash_sip",
			},
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(fmt.Sprintf("%s+%s", strings.Join(tt.targetTags, ","), strings.Join(tt.userTags, ",")), func(t *testing.T) {
			c := &Config{
				Target: &TargetSpec{
					BuildTags: tt.targetTags,
				},
				Options: &Options{
					Tags: tt.userTags,
				},
			}

			res := c.BuildTags()

			if len(res) != len(tt.result) {
				t.Errorf("expected %d tags, got %d", len(tt.result), len(res))
			}

			for i, tag := range tt.result {
				if tag != res[i] {
					t.Errorf("tag %d: expected %s, got %s", i, tt.result[i], tag)
				}
			}
		})
	}
}
