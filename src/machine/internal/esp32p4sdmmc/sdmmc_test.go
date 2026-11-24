package esp32p4sdmmc

import (
	"testing"
	"unsafe"
)

func TestCommandWord(t *testing.T) {
	tests := []struct {
		name string
		opts CommandOptions
		want []uint32 // bits that must be set
		not  []uint32 // bits that must be clear
	}{
		{
			name: "CMD0",
			opts: CommandOptions{Index: CmdGoIdleState},
			want: []uint32{1 << 15, 1 << 29, 1 << 31},
			not:  []uint32{1 << 13},
		},
		{
			name: "CMD8",
			opts: CommandOptions{Index: CmdSendIfCond, Response: RespR7, CheckCRC: true, CardNum: 0},
			want: []uint32{8, 1 << 6, 1 << 8, 1 << 13, 1 << 29, 1 << 31},
		},
		{
			name: "CMD17",
			opts: CommandOptions{Index: CmdReadSingleBlock, Response: RespR1, CheckCRC: true, Data: true},
			want: []uint32{17, 1 << 6, 1 << 8, 1 << 9, 1 << 13, 1 << 29, 1 << 31},
			not:  []uint32{1 << 10, 1 << 12},
		},
		{
			name: "CMD24",
			opts: CommandOptions{Index: CmdWriteSingleBlock, Response: RespR1, CheckCRC: true, Data: true, Write: true},
			want: []uint32{24, 1 << 6, 1 << 8, 1 << 9, 1 << 10, 1 << 13},
		},
		{
			name: "CMD18 auto-stop",
			opts: CommandOptions{Index: CmdReadMultipleBlock, Response: RespR1, Data: true, AutoStop: true},
			want: []uint32{18, 1 << 9, 1 << 12},
		},
		{
			name: "CMD12 stop",
			opts: CommandOptions{Index: CmdStopTransmission, Response: RespR1b},
			want: []uint32{12, 1 << 6, 1 << 14, 1 << 29, 1 << 31},
			not:  []uint32{1 << 13},
		},
		{
			name: "ACMD41",
			opts: CommandOptions{Index: CmdSendOpCond, Response: RespR3},
			want: []uint32{41, 1 << 6, 1 << 13, 1 << 29, 1 << 31},
		},
		{
			name: "update clock",
			opts: CommandOptions{Index: 0, UpdateClock: true},
			want: []uint32{1 << 21, 1 << 29, 1 << 31},
			not:  []uint32{1 << 13, 1 << 15},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.opts.Word()
			for _, bit := range tc.want {
				if got&bit == 0 {
					t.Errorf("word %#08x missing bit %#08x", got, bit)
				}
			}
			for _, bit := range tc.not {
				if got&bit != 0 {
					t.Errorf("word %#08x unexpectedly has bit %#08x", got, bit)
				}
			}
		})
	}
}

func TestClockDiv(t *testing.T) {
	tests := []struct {
		src, target   uint32
		host, cardDiv uint32
	}{
		{160_000_000, 400_000, 10, 20},
		{160_000_000, 500_000, 2, 80},
		{160_000_000, 20_000_000, 8, 0},
		{160_000_000, 25_000_000, 7, 0},
		{160_000_000, 40_000_000, 4, 0},
		{160_000_000, 1_000_000, 2, 40},
	}
	for _, tc := range tests {
		host, cardDiv := ClockDiv(tc.src, tc.target)
		if host != tc.host || cardDiv != tc.cardDiv {
			t.Errorf("ClockDiv(%d, %d) = (%d, %d), want (%d, %d)",
				tc.src, tc.target, host, cardDiv, tc.host, tc.cardDiv)
		}
	}
}

func TestClockEdge(t *testing.T) {
	h, n, l, update := ClockEdge(10)
	if h != 4 || n != 9 || l != 9 || !update {
		t.Fatalf("ClockEdge(10) = (%d,%d,%d,%v), want (4,9,9,true)", h, n, l, update)
	}
	h, n, l, update = ClockEdge(1)
	if h != 0 || n != 0 || l != 0 || update {
		t.Fatalf("ClockEdge(1) = (%d,%d,%d,%v), want (0,0,0,false)", h, n, l, update)
	}
}

func TestCSDv2(t *testing.T) {
	var csd [16]byte
	csd[0] = 0x40 // CSD version 2
	csd[5] = 9    // read block length = 512
	// c_size = 8191 -> (8191+1)*512 KiB = 4 GiB.
	csd[7] = 0x00
	csd[8] = 0x1f
	csd[9] = 0xff
	if got := CSDVersion(csd); got != 2 {
		t.Fatalf("CSDVersion = %d, want 2", got)
	}
	if got := CSDReadBlockLen(csd); got != 512 {
		t.Fatalf("CSDReadBlockLen = %d, want 512", got)
	}
	const want = int64(4 * 1024 * 1024 * 1024)
	if got := CSDCapacityBytes(csd); got != want {
		t.Fatalf("CSDCapacityBytes = %d, want %d", got, want)
	}
	if got := CSDBlockCount(csd); got != want/512 {
		t.Fatalf("CSDBlockCount = %d, want %d", got, want/512)
	}
}

func TestCSDv1(t *testing.T) {
	var csd [16]byte
	csd[0] = 0x00 // CSD version 1
	csd[5] = 9    // read block length = 512
	// c_size = 0x123, c_size_mult = 2.
	csd[6] = 0x00
	csd[7] = 0x48
	csd[8] = 0xc0
	csd[9] = 0x01
	csd[10] = 0x00
	if got := CSDVersion(csd); got != 1 {
		t.Fatalf("CSDVersion = %d, want 1", got)
	}
	const want = int64((0x123 + 1) * (1 << (2 + 2)) * 512)
	if got := CSDCapacityBytes(csd); got != want {
		t.Fatalf("CSDCapacityBytes = %d, want %d", got, want)
	}
	if got := CSDBlockCount(csd); got != want/512 {
		t.Fatalf("CSDBlockCount = %d, want %d", got, want/512)
	}
}

func TestResponseBytes(t *testing.T) {
	// RESP0 is the least significant word of the 128-bit response.
	resp := [4]uint32{0x03020100, 0x0f0e0d0c, 0x1b1a1918, 0x27262524}
	got := ResponseBytes(resp)
	want := [16]byte{
		0x27, 0x26, 0x25, 0x24,
		0x1b, 0x1a, 0x19, 0x18,
		0x0f, 0x0e, 0x0d, 0x0c,
		0x03, 0x02, 0x01, 0x00,
	}
	if got != want {
		t.Fatalf("ResponseBytes = %x, want %x", got, want)
	}
}

func TestR1(t *testing.T) {
	if R1HasError(R1ReadyForData | R1AppCmd) {
		t.Fatal("status-only R1 reported as error")
	}
	if !R1HasError(R1ReadyForData | 1<<31) {
		t.Fatal("R1 with bit 31 not reported as error")
	}
}

func TestDescriptor(t *testing.T) {
	if got := unsafe.Sizeof(Descriptor{}); got != DescriptorSize {
		t.Fatalf("descriptor size = %d, want %d", got, DescriptorSize)
	}
	var d Descriptor
	d.SetFirst(true)
	d.SetLast(true)
	d.SetOwned(true)
	d.SetBuffer(0x12345678, 512)
	d.SetNext(0x87654321)

	if d.Word[0]&DescFirst == 0 || d.Word[0]&DescLast == 0 || !d.Owned() {
		t.Fatalf("descriptor flags not set: %#08x", d.Word[0])
	}
	if d.Word[0]&DescSecondChained == 0 {
		t.Fatalf("second_address_chained not set: %#08x", d.Word[0])
	}
	if d.Word[1]&0x1fff != 512 {
		t.Fatalf("buffer size = %d, want 512", d.Word[1]&0x1fff)
	}
	if d.Word[2] != 0x12345678 {
		t.Fatalf("buffer ptr = %#08x", d.Word[2])
	}
	if d.Word[3] != 0x87654321 {
		t.Fatalf("next ptr = %#08x", d.Word[3])
	}

	d.SetOwned(false)
	if d.Owned() {
		t.Fatal("descriptor still owned after SetOwned(false)")
	}
	d.SetLast(false)
	if d.Word[0]&DescLast != 0 {
		t.Fatal("last descriptor bit still set")
	}
}
