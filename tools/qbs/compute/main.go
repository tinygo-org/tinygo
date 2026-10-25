//go:build tinygo

package main

import (
	"runtime"
	"strconv"

	"github.com/tinygo-org/tinygo/tools/qbs/cases"
	"github.com/tinygo-org/tinygo/tools/qbs/dev"
)

func main() {
	var tests []dev.Test
	for _, c := range cases.All() {
		c := c
		tests = append(tests, dev.Test{
			Name: c.Name,
			Run: func(t *dev.T) {
				if c.Heap > 0 {
					runtime.GC()
					var m runtime.MemStats
					runtime.ReadMemStats(&m)
					if int(m.HeapIdle) < c.Heap+c.Heap/2 {
						t.Skip("needs " + strconv.Itoa(c.Heap) + " bytes of heap, " + strconv.Itoa(int(m.HeapIdle)) + " free")
						return
					}
				}
				t.Val("result", c.Fn())
			},
		})
	}
	dev.Main("compute", tests)
}
