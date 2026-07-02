package builder

import (
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
)

func TestESPFlashSize(t *testing.T) {
	tests := []struct {
		name             string
		defaultSpeedSize uint8
		target           *compileopts.TargetSpec
		wantSpeedSize    uint8
	}{
		{
			name:             "keeps esp32c3 default",
			defaultSpeedSize: 0x1f,
			target:           &compileopts.TargetSpec{},
			wantSpeedSize:    0x1f,
		},
		{
			name:             "sets esp32c3 4MB size nibble",
			defaultSpeedSize: 0x1f,
			target:           &compileopts.TargetSpec{ESPFlashSize: "4MB"},
			wantSpeedSize:    0x2f,
		},
		{
			name:             "keeps esp32c6 default frequency nibble",
			defaultSpeedSize: 0x10,
			target:           &compileopts.TargetSpec{ESPFlashSize: "4MB"},
			wantSpeedSize:    0x20,
		},
		{
			name:             "normalizes lowercase size",
			defaultSpeedSize: 0x1f,
			target:           &compileopts.TargetSpec{ESPFlashSize: "8mb"},
			wantSpeedSize:    0x3f,
		},
		{
			name:             "allows nil target",
			defaultSpeedSize: 0x1f,
			target:           nil,
			wantSpeedSize:    0x1f,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spiSpeedSize, err := setESPFlashSize(test.defaultSpeedSize, test.target)
			if err != nil {
				t.Fatal(err)
			}
			if spiSpeedSize != test.wantSpeedSize {
				t.Fatalf("unexpected spi speed/size: got %#x, want %#x", spiSpeedSize, test.wantSpeedSize)
			}
		})
	}
}

func TestESPFlashSizeInvalid(t *testing.T) {
	_, err := setESPFlashSize(0x1f, &compileopts.TargetSpec{ESPFlashSize: "6MB"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "esp-flash-size") {
		t.Fatalf("unexpected error: got %q, want it to contain %q", err, "esp-flash-size")
	}
}
