package main

var checksum uint32

func init() {
	genTable()
	for _, entry := range pdTable {
		checksum += uint32(entry.hivco[0]) + uint32(entry.hivco[1])
		checksum += uint32(entry.lovco[0]) + uint32(entry.lovco[1])
	}
}

func main() {
	initial := pdTable
	zero := pdTable[0]
	for product := range pdTable {
		pdTable[product] = zero
	}
	genTable()
	if pdTable != initial {
		panic("runtime table differs from initial table")
	}
	for product, entry := range pdTable {
		want := zero
		if product != 0 {
			best := int64(255)
			for pd1 := uint8(1); pd1 <= 7; pd1++ {
				for pd2 := uint8(1); pd2 <= pd1; pd2++ {
					distance := abs(int64(pd1*pd2) - int64(product))
					if distance < best {
						best = distance
						want.hivco = [2]uint8{pd1, pd2}
					}
					if distance == best {
						want.lovco = [2]uint8{pd1, pd2}
					}
				}
			}
			pdTable[product] = zero
			genTableEntry(product)
			if pdTable[product] != want || pdTable != initial {
				panic("entry helper changed the result or another entry")
			}
		}
		if entry != want {
			println("product", product)
			panic("wrong post-divider entry")
		}
	}
	pdTable[49] = zero
	before := pdTable
	genTable()
	if pdTable != before {
		panic("generated-table guard changed the table")
	}
	println(checksum)
}
