//go:build !avr

package cases

import "fmt"

func init() {
	register("fmt", 0, func() string {
		r := rng(30)
		s := ""
		for i := 0; i < 20; i++ {
			v := int32(r.next())
			s += fmt.Sprintf("%d %x %08.3f %q %v %5.1f|", v, uint32(v), float64(v)/1000, fmt.Sprint(i), []int{i, i + 1}, float32(i)/3)
		}
		return fmt.Sprintf("%d %x", len(s), fnv64(s))
	})
	needsHeap(8 << 10)
}
