package uefi

import "testing"

func TestStringToCHAR16RoundTrip(t *testing.T) {
	input := "Hello, 世界"
	encoded := StringToCHAR16(input)

	if got := CHAR16ToString(encoded); got != input {
		t.Fatalf("round trip mismatch: got %q want %q", got, input)
	}
}

func TestStringToCHAR16UsesSurrogatePairs(t *testing.T) {
	input := "🙂"
	encoded := StringToCHAR16(input)
	want := []CHAR16{0xD83D, 0xDE42}

	if len(encoded) != len(want) {
		t.Fatalf("encoded length mismatch: got %d want %d", len(encoded), len(want))
	}
	for i := range want {
		if encoded[i] != want[i] {
			t.Fatalf("encoded[%d] mismatch: got %#x want %#x", i, encoded[i], want[i])
		}
	}
}

func TestStringToCHAR16ZTerminates(t *testing.T) {
	encoded := StringToCHAR16Z("abc")

	if len(encoded) != 4 {
		t.Fatalf("terminated length mismatch: got %d want 4", len(encoded))
	}
	if encoded[len(encoded)-1] != 0 {
		t.Fatalf("missing NUL terminator: got %#x", encoded[len(encoded)-1])
	}
	if got := CHAR16PtrToString(&encoded[0]); got != "abc" {
		t.Fatalf("pointer round trip mismatch: got %q want %q", got, "abc")
	}
}

func TestBytesToCHAR16RoundTrip(t *testing.T) {
	input := []byte("UEFI µ")
	encoded := BytesToCHAR16(input)

	got := CHAR16ToBytes(encoded)
	if string(got) != string(input) {
		t.Fatalf("byte round trip mismatch: got %q want %q", got, input)
	}
}

func TestCHAR16PtrHelpersHandleNil(t *testing.T) {
	if got := CHAR16PtrToString(nil); got != "" {
		t.Fatalf("CHAR16PtrToString(nil) = %q, want empty string", got)
	}
	if got := CHAR16PtrLenToString(nil, 3); got != "" {
		t.Fatalf("CHAR16PtrLenToString(nil, 3) = %q, want empty string", got)
	}
}
