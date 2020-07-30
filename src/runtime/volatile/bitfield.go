package volatile

type BitField struct{
	Lsb,Msb int
	Ptr *Register32
}

func (b *BitField) SetBits(v uint32) {
	mask:=uint32(0)
	for i:=b.Lsb; i<=b.Msb; i++ {
		mask|=(1 << i)
	}
 	b.Ptr.ReplaceBits(v,mask,uint8(b.Lsb))
}
func (b *BitField) Clear() {
	b.SetBits(0)
}

func (b *BitField) Set() {
	diff:=b.Msb-b.Lsb
	value:=(1<<(diff+1)) - 1 //all ffs, width of diff
	b.SetBits(uint32(value))
}

//Get returns the value of this bitfield, not the shifted representation in the reg
func (b *BitField) Get() uint32 {
	v:=b.Ptr.Get()
	v>>=b.Lsb
	diff:=b.Msb-b.Lsb
	value:=(1<<(diff+1)) - 1 //all ffs, width of diff
	return v&uint32(value)
}
func (b *BitField) HasBits() bool {
	diff:=b.Msb-b.Lsb
	value:=(1<<(diff+1)) - 1 //all ffs, width of diff
	value<<=b.Lsb
	return b.Ptr.HasBits(uint32(value))
}