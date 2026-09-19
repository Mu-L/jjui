package annotation

import (
	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/idursun/jjui/internal/ui/colorcontrast"
	"github.com/idursun/jjui/internal/ui/layout"
)

const (
	// Leave headroom for syntax colours instead of forcing every token to white
	// or black. The final foreground/background pair still needs 4.5:1.
	highlightTextContrast = 7.0
	syntaxTextContrast    = colorcontrast.Minimum
)

type contrastRGB = colorcontrast.RGB

func readableBackground(background, surface contrastRGB) contrastRGB {
	black, white := contrastRGB{0, 0, 0}, contrastRGB{255, 255, 255}
	edge := white
	if colorcontrast.Ratio(black, surface) > colorcontrast.Ratio(white, surface) {
		edge = black
	}
	if colorcontrast.Ratio(edge, background) >= highlightTextContrast ||
		colorcontrast.Ratio(edge, surface) < highlightTextContrast {
		return background
	}
	return colorcontrast.Toward(background, surface, edge, highlightTextContrast)
}

type contrastPair struct {
	foreground ansi.Color
	background ansi.Color
}

// contrastAdjuster is shared by the visible rows for one frame. Cache colour
// pairs so contrast searches are per style, not per character. A fresh cache
// also avoids retaining stale results after theme or terminal palette changes.
type contrastAdjuster struct {
	surface         contrastRGB
	terminalPalette map[int]string
	pairs           map[contrastPair]contrastPair
}

func newContrastAdjuster(surface ansi.Color, terminalBackground string, terminalPalette map[int]string, dark bool) *contrastAdjuster {
	a := &contrastAdjuster{terminalPalette: terminalPalette, pairs: make(map[contrastPair]contrastPair)}
	if !dark {
		a.surface = contrastRGB{255, 255, 255}
	}
	if terminalBackground != "" {
		a.surface = a.rgb(lipgloss.Color(terminalBackground))
	}
	if surface != nil {
		if _, unset := surface.(lipgloss.NoColor); !unset {
			a.surface = a.rgb(surface)
		}
	}
	return a
}

func (a *contrastAdjuster) rgb(value ansi.Color) contrastRGB {
	c := syntaxColour(value, a.terminalPalette)
	return contrastRGB{c.Red(), c.Green(), c.Blue()}
}

func (a *contrastAdjuster) adjust(pair contrastPair) contrastPair {
	if result, ok := a.pairs[pair]; ok {
		return result
	}
	result := pair
	bg := a.rgb(pair.background)
	adjustedBG := readableBackground(bg, a.surface)
	if adjustedBG != bg {
		result.background = adjustedBG.Color()
	}
	if pair.foreground != nil {
		fg := a.rgb(pair.foreground)
		adjustedFG := colorcontrast.Foreground(fg, adjustedBG)
		if adjustedFG != fg {
			result.foreground = adjustedFG.Color()
		}
	}
	a.pairs[pair] = result
	return result
}

type annotationContrastEffect struct {
	rect     layout.Rectangle
	adjuster *contrastAdjuster
}

func (e annotationContrastEffect) GetRect() layout.Rectangle { return e.rect }
func (e annotationContrastEffect) GetZ() int                 { return 3 }

// Apply runs after word, line and selection backgrounds have been composed.
// Work on screen cells so ANSI resets and token boundaries cannot undo it.
func (e annotationContrastEffect) Apply(buf uv.Screen) {
	bounds := e.rect.Intersect(buf.Bounds())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; {
			cell := buf.CellAt(x, y)
			if cell == nil || cell.Width == 0 {
				x++
				continue
			}
			x += cell.Width
			pair := contrastPair{cell.Style.Fg, cell.Style.Bg}
			reverse := cell.Style.Attrs&uv.AttrReverse != 0
			if reverse {
				pair.foreground, pair.background = pair.background, pair.foreground
			}
			if pair.background == nil {
				continue
			}
			if _, unset := pair.background.(lipgloss.NoColor); unset {
				continue
			}
			adjusted := e.adjuster.adjust(pair)
			if adjusted == pair {
				continue
			}
			updated := cell.Clone()
			if reverse {
				adjusted.foreground, adjusted.background = adjusted.background, adjusted.foreground
			}
			updated.Style.Fg, updated.Style.Bg = adjusted.foreground, adjusted.background
			buf.SetCell(x-cell.Width, y, updated)
		}
	}
}
