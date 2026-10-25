package cases

import (
	"math/bits"
	"sort"
	"strconv"
)

func init() {
	register("int-arith", 0, func() string {
		r := rng(1)
		acc := uint64(1469598103934665603)
		for i := 0; i < 2000; i++ {
			a := int64(int32(r.next())) * int64(int32(r.next()))
			b := int64(r.next()%1000) + 1
			u := uint64(r.next())<<32 | uint64(r.next())
			v := uint64(r.next()) | 1
			acc = acc*1099511628211 ^ uint64(a)
			acc ^= uint64(a / b)
			acc ^= uint64(a % b)
			acc ^= u / v
			acc ^= u % v
			acc += uint64(int32(r.next()) >> uint(i%31))
			acc ^= u<<uint(i%64) + u>>uint(i%64)
		}
		return hex64(acc)
	})

	register("bits", 0, func() string {
		r := rng(2)
		var acc uint64
		for i := 0; i < 1000; i++ {
			x := uint64(r.next())<<32 | uint64(r.next())
			y := uint32(r.next())
			hi, lo := bits.Mul64(x, uint64(y)|1)
			sum, carry := bits.Add64(x, lo, uint64(i&1))
			acc += uint64(bits.LeadingZeros64(x)) + uint64(bits.TrailingZeros32(y))
			acc += uint64(bits.OnesCount64(x)) + uint64(bits.Len32(y))
			acc ^= bits.RotateLeft64(x, i%64) ^ uint64(bits.Reverse32(y)) ^ hi ^ sum ^ carry
		}
		return hex64(acc)
	})

	register("sort", 0, func() string {
		r := rng(3)
		ints := make([]int, 400)
		for i := range ints {
			ints[i] = int(int32(r.next()))
		}
		sort.Ints(ints)
		var acc uint64
		for i, v := range ints {
			acc = acc*31 + uint64(v)*uint64(i+1)
		}
		strs := make([]string, 60)
		for i := range strs {
			strs[i] = strconv.Itoa(int(r.next() % 100000))
		}
		sort.Strings(strs)
		for _, s := range strs {
			acc = acc*131 + fnv64(s)
		}
		type pair struct{ k, v int }
		pairs := make([]pair, 100)
		for i := range pairs {
			pairs[i] = pair{int(r.next() % 50), i}
		}
		sort.SliceStable(pairs, func(i, j int) bool { return pairs[i].k < pairs[j].k })
		for _, p := range pairs {
			acc = acc*7 + uint64(p.k*1000+p.v)
		}
		return hex64(acc)
	})
	needsHeap(8 << 10)

	register("map", 0, func() string {
		r := rng(4)
		m := make(map[uint32]uint32)
		for i := 0; i < 300; i++ {
			m[r.next()%500] = r.next()
		}
		for k := uint32(0); k < 500; k += 3 {
			delete(m, k)
		}
		keys := make([]int, 0, len(m))
		for k := range m {
			keys = append(keys, int(k))
		}
		sort.Ints(keys)
		var acc uint64
		for _, k := range keys {
			acc = acc*1000003 ^ uint64(k)<<32 ^ uint64(m[uint32(k)])
		}
		return strconv.Itoa(len(m)) + " " + hex64(acc)
	})
}
