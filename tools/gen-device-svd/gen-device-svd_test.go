package main

import "testing"

func TestProcessClusterSizeDim1(t *testing.T) {
	dim := 1
	off0, off4 := "0x0", "0x4"
	p := &Peripheral{Name: "TIMG", BaseAddress: 0x1000}
	processCluster(p, []*SVDCluster{
		{
			Name:          "T%s",
			Dim:           &dim,
			AddressOffset: "0x0",
			Registers: []*SVDRegister{
				{Name: "CONFIG", AddressOffset: &off0},
				{Name: "LO", AddressOffset: &off4},
			},
		},
	}, map[string]*Peripheral{})

	if len(p.Registers) != 1 {
		t.Fatalf("got %d registers, want 1", len(p.Registers))
	}
	if got := p.Registers[0].ElementSize; got != 8 {
		t.Errorf("cluster ElementSize = %d, want 8", got)
	}
}
