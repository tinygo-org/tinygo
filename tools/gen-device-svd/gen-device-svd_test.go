package main

import "testing"

func TestParseBitfieldsDimensionedSingleBit(t *testing.T) {
	dim := "2"
	dimIndex := "0-1"
	lsb := uint32(0)
	msb := uint32(0)
	fields, bitfields := parseBitfields("ADC", "CHSELR0", []*SVDField{
		{
			Name:         "CHSEL%s",
			Dim:          &dim,
			DimIndex:     &dimIndex,
			DimIncrement: "1",
			Lsb:          &lsb,
			Msb:          &msb,
		},
	}, "")

	wantBitfields := []Bitfield{
		{Name: "CHSEL0", Offset: 0, Mask: 1},
		{Name: "CHSEL1", Offset: 1, Mask: 2},
	}
	if len(bitfields) != len(wantBitfields) {
		t.Fatalf("got %d bitfields, want %d", len(bitfields), len(wantBitfields))
	}
	for i, want := range wantBitfields {
		if got := bitfields[i]; got != want {
			t.Errorf("bitfield %d = %+v, want %+v", i, got, want)
		}
	}

	wantConstants := map[string]uint64{
		"ADC_CHSELR0_CHSEL0_Pos": 0,
		"ADC_CHSELR0_CHSEL0_Msk": 1,
		"ADC_CHSELR0_CHSEL0":     1,
		"ADC_CHSELR0_CHSEL1_Pos": 1,
		"ADC_CHSELR0_CHSEL1_Msk": 2,
		"ADC_CHSELR0_CHSEL1":     2,
	}
	if len(fields) != len(wantConstants) {
		t.Fatalf("got %d constants, want %d", len(fields), len(wantConstants))
	}
	for _, field := range fields {
		want, ok := wantConstants[field.Name]
		if !ok {
			t.Errorf("unexpected constant %q", field.Name)
			continue
		}
		if field.Value != want {
			t.Errorf("%s = %#x, want %#x", field.Name, field.Value, want)
		}
		delete(wantConstants, field.Name)
	}
	for name := range wantConstants {
		t.Errorf("missing constant %q", name)
	}
}
