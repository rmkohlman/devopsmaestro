package palette

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// HSL represents a color in the HSL (Hue, Saturation, Lightness) color space.
type HSL struct {
	H float64 // Hue: 0-360 degrees
	S float64 // Saturation: 0-1 (0% to 100%)
	L float64 // Lightness: 0-1 (0% to 100%)
}

// HexToHSL converts a hex color string to HSL values.
func HexToHSL(hex string) (HSL, error) {
	r, g, b, err := HexToRGB(hex)
	if err != nil {
		return HSL{}, err
	}
	return RGBToHSL(r, g, b), nil
}

// HexToRGB converts a hex color string to RGB values (0-255).
func HexToRGB(hex string) (r, g, b uint8, err error) {
	ri, gi, bi, err := ParseRGB(hex)
	if err != nil {
		return 0, 0, 0, err
	}
	return uint8(ri), uint8(gi), uint8(bi), nil
}

// RGBToHSL converts RGB values (0-255) to HSL.
func RGBToHSL(r, g, b uint8) HSL {
	rNorm := float64(r) / 255.0
	gNorm := float64(g) / 255.0
	bNorm := float64(b) / 255.0

	max := math.Max(rNorm, math.Max(gNorm, bNorm))
	min := math.Min(rNorm, math.Min(gNorm, bNorm))

	var h, s, l float64
	l = (max + min) / 2.0

	if max == min {
		// Achromatic (gray)
		h = 0
		s = 0
	} else {
		diff := max - min

		if l > 0.5 {
			s = diff / (2.0 - max - min)
		} else {
			s = diff / (max + min)
		}

		switch max {
		case rNorm:
			h = (gNorm-bNorm)/diff + func() float64 {
				if gNorm < bNorm {
					return 6
				}
				return 0
			}()
		case gNorm:
			h = (bNorm-rNorm)/diff + 2
		case bNorm:
			h = (rNorm-gNorm)/diff + 4
		}

		h *= 60
	}

	// Ensure hue is in 0-360 range
	if h < 0 {
		h += 360
	}

	return HSL{H: h, S: s, L: l}
}

// ToHex converts HSL values to a hex color string.
func (h HSL) ToHex() string {
	r, g, b := h.ToRGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// ToRGB converts HSL values to RGB (0-255).
func (h HSL) ToRGB() (r, g, b uint8) {
	if h.S == 0 {
		// Achromatic
		val := uint8(h.L * 255)
		return val, val, val
	}

	var q float64
	if h.L < 0.5 {
		q = h.L * (1 + h.S)
	} else {
		q = h.L + h.S - h.L*h.S
	}

	p := 2*h.L - q
	hNorm := h.H / 360.0

	rNorm := hueToRGB(p, q, hNorm+1.0/3.0)
	gNorm := hueToRGB(p, q, hNorm)
	bNorm := hueToRGB(p, q, hNorm-1.0/3.0)

	return uint8(rNorm * 255), uint8(gNorm * 255), uint8(bNorm * 255)
}

// hueToRGB is a helper function for HSL to RGB conversion.
func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

// Shift returns a new HSL color with the specified deltas applied.
func (h HSL) Shift(hDelta, sDelta, lDelta float64) HSL {
	newH := h.H + hDelta
	newS := h.S + sDelta
	newL := h.L + lDelta

	// Normalize hue to 0-360
	for newH < 0 {
		newH += 360
	}
	for newH >= 360 {
		newH -= 360
	}

	// Clamp saturation and lightness to 0-1
	if newS < 0 {
		newS = 0
	}
	if newS > 1 {
		newS = 1
	}
	if newL < 0 {
		newL = 0
	}
	if newL > 1 {
		newL = 1
	}

	return HSL{H: newH, S: newS, L: newL}
}

// WithHue returns a new HSL color with the specified hue.
func (h HSL) WithHue(hue float64) HSL {
	// Normalize hue to 0-360
	for hue < 0 {
		hue += 360
	}
	for hue >= 360 {
		hue -= 360
	}

	return HSL{H: hue, S: h.S, L: h.L}
}

// WithSaturation returns a new HSL color with the specified saturation.
func (h HSL) WithSaturation(saturation float64) HSL {
	// Clamp to 0-1
	if saturation < 0 {
		saturation = 0
	}
	if saturation > 1 {
		saturation = 1
	}

	return HSL{H: h.H, S: saturation, L: h.L}
}

// WithLightness returns a new HSL color with the specified lightness.
func (h HSL) WithLightness(lightness float64) HSL {
	// Clamp to 0-1
	if lightness < 0 {
		lightness = 0
	}
	if lightness > 1 {
		lightness = 1
	}

	return HSL{H: h.H, S: h.S, L: lightness}
}

// String returns a string representation of the HSL color.
func (h HSL) String() string {
	return fmt.Sprintf("hsl(%.1f, %.1f%%, %.1f%%)", h.H, h.S*100, h.L*100)
}

// ParseHSL parses a HSL string like "hsl(210, 50%, 30%)" or "210,0.5,0.3".
func ParseHSL(s string) (HSL, error) {
	s = strings.TrimSpace(s)

	// Handle "hsl(h, s%, l%)" format
	if strings.HasPrefix(s, "hsl(") && strings.HasSuffix(s, ")") {
		s = s[4 : len(s)-1] // Remove "hsl(" and ")"
		parts := strings.Split(s, ",")
		if len(parts) != 3 {
			return HSL{}, fmt.Errorf("invalid HSL format: expected 3 components")
		}

		h, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return HSL{}, fmt.Errorf("invalid hue: %w", err)
		}

		sStr := strings.TrimSpace(strings.TrimSuffix(parts[1], "%"))
		s, err := strconv.ParseFloat(sStr, 64)
		if err != nil {
			return HSL{}, fmt.Errorf("invalid saturation: %w", err)
		}
		if strings.HasSuffix(parts[1], "%") {
			s /= 100.0
		}

		lStr := strings.TrimSpace(strings.TrimSuffix(parts[2], "%"))
		l, err := strconv.ParseFloat(lStr, 64)
		if err != nil {
			return HSL{}, fmt.Errorf("invalid lightness: %w", err)
		}
		if strings.HasSuffix(parts[2], "%") {
			l /= 100.0
		}

		return HSL{H: h, S: s, L: l}, nil
	}

	// Handle "h,s,l" format (comma-separated, s and l as decimals)
	parts := strings.Split(s, ",")
	if len(parts) == 3 {
		h, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return HSL{}, fmt.Errorf("invalid hue: %w", err)
		}

		s, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return HSL{}, fmt.Errorf("invalid saturation: %w", err)
		}

		l, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
		if err != nil {
			return HSL{}, fmt.Errorf("invalid lightness: %w", err)
		}

		return HSL{H: h, S: s, L: l}, nil
	}

	return HSL{}, fmt.Errorf("invalid HSL format: %s", s)
}

// Lighten returns a new HSL color with increased lightness.
func (h HSL) Lighten(amount float64) HSL {
	return h.Shift(0, 0, amount)
}

// Darken returns a new HSL color with decreased lightness.
func (h HSL) Darken(amount float64) HSL {
	return h.Shift(0, 0, -amount)
}

// Saturate returns a new HSL color with increased saturation.
func (h HSL) Saturate(amount float64) HSL {
	return h.Shift(0, amount, 0)
}

// Desaturate returns a new HSL color with decreased saturation.
func (h HSL) Desaturate(amount float64) HSL {
	return h.Shift(0, -amount, 0)
}

// Rotate returns a new HSL color with rotated hue.
func (h HSL) Rotate(degrees float64) HSL {
	return h.Shift(degrees, 0, 0)
}
