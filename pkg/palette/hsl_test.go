package palette

import (
	"strings"
	"testing"
)

func TestHSLConversions(t *testing.T) {
	testCases := []struct {
		hex      string
		expected HSL
	}{
		{"#FF0000", HSL{H: 0, S: 1.0, L: 0.5}},   // Pure red
		{"#00FF00", HSL{H: 120, S: 1.0, L: 0.5}}, // Pure green
		{"#0000FF", HSL{H: 240, S: 1.0, L: 0.5}}, // Pure blue
		{"#FFFFFF", HSL{H: 0, S: 0.0, L: 1.0}},   // White
		{"#000000", HSL{H: 0, S: 0.0, L: 0.0}},   // Black
	}

	for _, tc := range testCases {
		hsl, err := HexToHSL(tc.hex)
		if err != nil {
			t.Errorf("HexToHSL(%s) failed: %v", tc.hex, err)
			continue
		}

		// Allow small floating point differences
		if abs(hsl.H-tc.expected.H) > 1.0 {
			t.Errorf("HexToHSL(%s) hue: expected %.1f, got %.1f", tc.hex, tc.expected.H, hsl.H)
		}
		if abs(hsl.S-tc.expected.S) > 0.01 {
			t.Errorf("HexToHSL(%s) saturation: expected %.2f, got %.2f", tc.hex, tc.expected.S, hsl.S)
		}
		if abs(hsl.L-tc.expected.L) > 0.01 {
			t.Errorf("HexToHSL(%s) lightness: expected %.2f, got %.2f", tc.hex, tc.expected.L, hsl.L)
		}

		// Test round-trip conversion (case-insensitive comparison)
		backToHex := hsl.ToHex()
		if !strings.EqualFold(backToHex, tc.hex) {
			t.Errorf("Round-trip failed: %s -> %s -> %s", tc.hex, hsl.String(), backToHex)
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestHSLShifting(t *testing.T) {
	// Start with blue
	blue := HSL{H: 240, S: 1.0, L: 0.5}

	// Shift to red (rotate -120°)
	red := blue.Shift(-120, 0, 0)
	expected := HSL{H: 120, S: 1.0, L: 0.5} // Should be green

	if abs(red.H-expected.H) > 1.0 {
		t.Errorf("Hue shift failed: expected %.1f, got %.1f", expected.H, red.H)
	}

	// Test hue wrapping
	wrapped := blue.Shift(150, 0, 0) // 240 + 150 = 390, should wrap to 30
	if abs(wrapped.H-30) > 1.0 {
		t.Errorf("Hue wrapping failed: expected 30, got %.1f", wrapped.H)
	}
}

func TestCoolNightColors(t *testing.T) {
	// Test the CoolNight Ocean reference colors
	coolNightBg := "#011628"

	hsl, err := HexToHSL(coolNightBg)
	if err != nil {
		t.Fatalf("Failed to parse CoolNight bg: %v", err)
	}

	// Should be around blue hue (210°)
	if hsl.H < 200 || hsl.H > 220 {
		t.Errorf("CoolNight bg hue expected ~210°, got %.1f°", hsl.H)
	}

	// Should be very dark (low lightness)
	if hsl.L > 0.1 {
		t.Errorf("CoolNight bg should be very dark, got lightness %.2f", hsl.L)
	}
}
