package common

import (
	"image/color"

	"github.com/charmbracelet/x/ansi"
	"github.com/idursun/jjui/internal/ui/colorcontrast"
	"github.com/idursun/jjui/internal/ui/render"
)

// NewForegroundContrast uses known terminal colours only. In particular, an
// implicit default foreground must not be guessed from light/dark mode.
func (p *Palette) NewForegroundContrast() *render.ForegroundContrast {
	return render.NewForegroundContrast(func(value color.Color) (colorcontrast.RGB, bool) {
		if value == nil {
			return colorcontrast.RGB{}, false
		}
		hex, ok := ResolveTerminalColor(value, p.blend.terminalPalette)
		if !ok {
			return colorcontrast.RGB{}, false
		}
		parsed := ansi.XParseColor(hex)
		if parsed == nil {
			return colorcontrast.RGB{}, false
		}
		r, g, b, _ := parsed.RGBA()
		return colorcontrast.RGB{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}, true
	})
}
