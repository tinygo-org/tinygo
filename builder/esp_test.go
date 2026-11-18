package builder

import (
	"bytes"
	"testing"
)

func TestESP32P4FlashAddressClassification(t *testing.T) {
	cases := []struct {
		addr uint32
		want bool
	}{
		{0x3FFFFFFF, false},
		{0x40000000, true},
		{0x40010000, true},
		{0x43FFFFFF, true},
		{0x44000000, false},
		{0x48000000, false}, // PSRAM
		{0x4FF40000, false}, // SRAM
	}
	for _, c := range cases {
		if got := isESP32P4FlashAddress(c.addr); got != c.want {
			t.Errorf("isESP32P4FlashAddress(0x%x) = %v, want %v", c.addr, got, c.want)
		}
	}
}

func TestESP32P4FlashImageOffset(t *testing.T) {
	cases := []struct {
		addr uint32
		want int
		ok   bool
	}{
		{0x40010000, 0xE000, true},
		{0x4002F128, 0x2D128, true},
		{0x40000000, 0, false},
		{0x40001FFF, 0, false},
		{0x44000000, 0, false},
		{0x4FF40000, 0, false},
	}
	for _, c := range cases {
		got, ok := esp32p4FlashImageOffset(c.addr)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("esp32p4FlashImageOffset(0x%x) = 0x%x, %v; want 0x%x, %v", c.addr, got, ok, c.want, c.ok)
		}
	}
}

func TestAppendESP32P4FlashSegmentsPadsVMAGaps(t *testing.T) {
	var out bytes.Buffer
	segs := []*espImageSegment{
		{addr: 0x40010000, data: []byte{1, 2, 3, 4}},
		{addr: 0x40010010, data: []byte{5, 6}},
		{addr: 0x4002F128, data: []byte{7, 8}},
	}
	if err := appendESP32P4FlashSegments(&out, segs); err != nil {
		t.Fatal(err)
	}

	b := out.Bytes()
	if want := 0x2D128 + 2; len(b) != want {
		t.Fatalf("image length = 0x%x, want 0x%x", len(b), want)
	}
	if b[0xE000] != 1 || b[0xE001] != 2 || b[0xE002] != 3 || b[0xE003] != 4 {
		t.Fatalf("first segment not at 0xE000: %v", b[0xE000:0xE004])
	}
	if b[0xE004] != 0 {
		t.Fatalf("gap after first segment is not zero: %v", b[0xE000:0xE010])
	}
	if b[0xE010] != 5 || b[0xE011] != 6 {
		t.Fatalf("second segment not at 0xE010: %v", b[0xE010:0xE012])
	}
	if b[0x2D128] != 7 || b[0x2D129] != 8 {
		t.Fatalf("third segment not at 0x2D128: %v", b[0x2D128:0x2D12A])
	}
}

func TestAppendESP32P4FlashSegmentsSortsByAddress(t *testing.T) {
	var out bytes.Buffer
	segs := []*espImageSegment{
		{addr: 0x40010010, data: []byte{5, 6}},
		{addr: 0x40010000, data: []byte{1, 2, 3, 4}},
	}
	if err := appendESP32P4FlashSegments(&out, segs); err != nil {
		t.Fatal(err)
	}

	b := out.Bytes()
	if b[0xE000] != 1 || b[0xE001] != 2 || b[0xE002] != 3 || b[0xE003] != 4 {
		t.Fatalf("lowest segment not written first: %v", b[0xE000:0xE004])
	}
	if b[0xE010] != 5 || b[0xE011] != 6 {
		t.Fatalf("higher segment not at 0xE010: %v", b[0xE010:0xE012])
	}
}

func TestAppendESP32P4FlashSegmentsRejectsOverlap(t *testing.T) {
	var out bytes.Buffer
	segs := []*espImageSegment{
		{addr: 0x40010000, data: []byte{1, 2, 3, 4}},
		{addr: 0x40010002, data: []byte{5, 6}},
	}
	if err := appendESP32P4FlashSegments(&out, segs); err == nil {
		t.Fatal("expected overlap error")
	}
}

func TestAppendESP32P4FlashSegmentsRejectsOutsideWindow(t *testing.T) {
	var out bytes.Buffer
	segs := []*espImageSegment{
		{addr: 0x4FF40000, data: []byte{1, 2, 3, 4}},
	}
	if err := appendESP32P4FlashSegments(&out, segs); err == nil {
		t.Fatal("expected error for segment outside the flash window")
	}
}
