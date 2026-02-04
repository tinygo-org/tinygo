//go:build uefi

package runtime

import "unsafe"

// =============================================================================
// Minimal libc replacements
// =============================================================================
//
// These are required because LLVM lowers operations like struct zeroing and
// slice copying into calls to memset/memcpy/memmove. fmt package needs log()
// for float formatting.

// maxArraySize is the maximum array size for pointer-to-array casts.
// On 64-bit: 1<<48-1 (256 TiB - 1), on 32-bit: 1<<31-1 (2 GiB - 1).
const maxArraySize = (1 << (31 + 17*(^uintptr(0)>>63))) - 1

//export memset
func libc_memset(dest unsafe.Pointer, c int, n uintptr) unsafe.Pointer {
	d := (*[maxArraySize]byte)(dest)
	val := byte(c)
	for i := uintptr(0); i < n; i++ {
		d[i] = val
	}
	return dest
}

//export memcpy
func libc_memcpy(dest, src unsafe.Pointer, n uintptr) unsafe.Pointer {
	d := (*[maxArraySize]byte)(dest)
	s := (*[maxArraySize]byte)(src)
	for i := uintptr(0); i < n; i++ {
		d[i] = s[i]
	}
	return dest
}

//export memmove
func libc_memmove(dest, src unsafe.Pointer, n uintptr) unsafe.Pointer {
	if uintptr(dest) < uintptr(src) {
		// Copy forward.
		d := (*[maxArraySize]byte)(dest)
		s := (*[maxArraySize]byte)(src)
		for i := uintptr(0); i < n; i++ {
			d[i] = s[i]
		}
	} else {
		// Copy backward (handles overlapping regions where dest > src).
		d := (*[maxArraySize]byte)(dest)
		s := (*[maxArraySize]byte)(src)
		for i := n; i > 0; i-- {
			d[i-1] = s[i-1]
		}
	}
	return dest
}

//export memcmp
func libc_memcmp(s1, s2 unsafe.Pointer, n uintptr) int {
	p1 := (*[maxArraySize]byte)(s1)
	p2 := (*[maxArraySize]byte)(s2)
	for i := uintptr(0); i < n; i++ {
		if p1[i] != p2[i] {
			return int(p1[i] - p2[i])
		}
	}
	return 0
}

//export strlen
func libc_strlen(s unsafe.Pointer) uintptr {
	p := (*[maxArraySize]byte)(s)
	var n uintptr
	for p[n] != 0 {
		n++
	}
	return n
}

// log implements the natural logarithm needed by fmt for float formatting.
// Ported from musl libc.
//
//export log
func libc_log(x float64) float64 {
	const (
		ln2Hi = 6.93147180369123816490e-01
		ln2Lo = 1.90821492927058770002e-10
		lg1   = 6.666666666666735130e-01
		lg2   = 3.999999999940941908e-01
		lg3   = 2.857142874366239149e-01
		lg4   = 2.222219843214978396e-01
		lg5   = 1.818357216161805012e-01
		lg6   = 1.531383769920937332e-01
		lg7   = 1.479819860511658591e-01
	)

	u := *(*uint64)(unsafe.Pointer(&x))
	hx := uint32(u >> 32)
	k := 0

	if hx < 0x00100000 || hx>>31 != 0 {
		if u<<1 == 0 {
			return -1 / (x * x) // log(+-0) = -inf
		}
		if hx>>31 != 0 {
			return (x - x) / 0.0 // log(-x) = NaN
		}
		k -= 54
		x *= 0x1p54
		u = *(*uint64)(unsafe.Pointer(&x))
		hx = uint32(u >> 32)
	} else if hx >= 0x7ff00000 {
		return x
	} else if hx == 0x3ff00000 && u<<32 == 0 {
		return 0
	}

	hx += 0x3ff00000 - 0x3fe6a09e
	k += int(hx>>20) - 0x3ff
	hx = (hx & 0x000fffff) + 0x3fe6a09e
	u = uint64(hx)<<32 | (u & 0xffffffff)
	x = *(*float64)(unsafe.Pointer(&u))

	f := x - 1.0
	hfsq := 0.5 * f * f
	s := f / (2.0 + f)
	z := s * s
	w := z * z
	t1 := w * (lg2 + w*(lg4+w*lg6))
	t2 := z * (lg1 + w*(lg3+w*(lg5+w*lg7)))
	R := t2 + t1
	dk := float64(k)
	return s*(hfsq+R) + dk*ln2Lo - hfsq + f + dk*ln2Hi
}
