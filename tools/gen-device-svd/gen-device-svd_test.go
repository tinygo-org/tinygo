package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestWriteGoSharedInterrupt(t *testing.T) {
	device := &Device{
		Metadata: &Metadata{Name: "TEST", NameLower: "test"},
		Interrupts: []*Interrupt{
			{Name: "APB_ADC", HandlerName: "APB_ADC_IRQHandler", Value: 71},
			{Name: "TEMPERATURE_SENSOR", HandlerName: "TEMPERATURE_SENSOR_IRQHandler", Value: 71},
			{Name: "MCPWM0", HandlerName: "MCPWM0_IRQHandler", Value: 72},
		},
	}
	dir := t.TempDir()
	if err := writeGo(dir, device, "software"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "test.go"))
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)

	for _, name := range []string{"APB_ADC", "TEMPERATURE_SENSOR", "MCPWM0"} {
		if !strings.Contains(out, "IRQ_"+name+" = ") {
			t.Errorf("missing constant IRQ_%s", name)
		}
	}
	if got := strings.Count(out, "\tcase IRQ_"); got != 2 {
		t.Errorf("got %d cases in HandleInterrupt, want 2", got)
	}
	if strings.Contains(out, "case IRQ_TEMPERATURE_SENSOR:") {
		t.Errorf("HandleInterrupt has a second case for interrupt 71")
	}
}
