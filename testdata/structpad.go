package main

// A struct with a nonzero size that ends in a zero-sized field is padded by a
// byte, as gc pads it, so that a pointer to that field does not point past the
// end of the object (https://go.dev/issue/9401).

import "unsafe"

type endsEmpty struct {
	data [4096]byte
	end  struct{}
}

type endsZeroArray struct {
	b   byte
	end [0]byte
}

type onlyEmpty struct {
	a struct{}
	b [0]byte
}

type nested struct {
	inner endsEmpty
	end   struct{}
}

type emptyFirst struct {
	end  struct{}
	data [3]byte
}

func main() {
	println("endsEmpty:", int(unsafe.Sizeof(endsEmpty{})))
	println("endsZeroArray:", int(unsafe.Sizeof(endsZeroArray{})))
	println("onlyEmpty:", int(unsafe.Sizeof(onlyEmpty{})))
	println("nested:", int(unsafe.Sizeof(nested{})))
	println("emptyFirst:", int(unsafe.Sizeof(emptyFirst{})))
	println("array of endsEmpty:", int(unsafe.Sizeof([2]endsEmpty{})))

	// The address of the final field must stay inside the object.
	p := new(endsEmpty)
	start := uintptr(unsafe.Pointer(p))
	end := uintptr(unsafe.Pointer(&p.end))
	println("end field inside object:", end-start < unsafe.Sizeof(*p))
}
