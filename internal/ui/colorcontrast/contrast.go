// Package colorcontrast adjusts resolved RGB colours for readable text.
package colorcontrast

import (
	"math"

	"github.com/charmbracelet/x/ansi"
)

// Minimum is the contrast target for ordinary text.
const Minimum = 4.5

// RGB is an opaque, resolved sRGB colour.
type RGB [3]uint8

// Luminance returns relative luminance in linear light.
func (c RGB) Luminance() float64 {
	var linear [3]float64
	for i, v := range c {
		s := float64(v) / 255
		if s <= 0.04045 {
			linear[i] = s / 12.92
		} else {
			linear[i] = math.Pow((s+0.055)/1.055, 2.4)
		}
	}
	return 0.2126*linear[0] + 0.7152*linear[1] + 0.0722*linear[2]
}

// Ratio returns the contrast ratio of two resolved colours.
func Ratio(a, b RGB) float64 {
	l1, l2 := a.Luminance(), b.Luminance()
	return (max(l1, l2) + 0.05) / (min(l1, l2) + 0.05)
}

func (c RGB) mix(target RGB, amount float64) RGB {
	var result RGB
	for i := range result {
		result[i] = uint8(math.Round(float64(c[i]) + (float64(target[i])-float64(c[i]))*amount))
	}
	return result
}

// Toward finds the smallest tint or shade that reaches the target.
// The endpoint must satisfy the ratio. Search the rounded, displayed colours,
// not just floating-point values which may round below the requested contrast.
func Toward(original, target, against RGB, ratio float64) RGB {
	low, high := 0.0, 1.0
	for range 16 {
		mid := (low + high) / 2
		if Ratio(original.mix(target, mid), against) >= ratio {
			high = mid
		} else {
			low = mid
		}
	}
	return original.mix(target, high)
}

// Foreground returns the nearest tint or shade meeting Minimum contrast.
// Readable colours are returned unchanged.
func Foreground(foreground, background RGB) RGB {
	if Ratio(foreground, background) >= Minimum {
		return foreground
	}
	var best RGB
	bestDistance := math.Inf(1)
	for _, edge := range []RGB{{0, 0, 0}, {255, 255, 255}} {
		if Ratio(edge, background) < Minimum {
			continue
		}
		candidate := Toward(foreground, edge, background, Minimum)
		distance := 0.0
		for i := range candidate {
			delta := float64(candidate[i]) - float64(foreground[i])
			distance += delta * delta
		}
		if distance < bestDistance {
			best, bestDistance = candidate, distance
		}
	}
	return best
}

func (c RGB) Color() ansi.RGBColor { return ansi.RGBColor{R: c[0], G: c[1], B: c[2]} }
