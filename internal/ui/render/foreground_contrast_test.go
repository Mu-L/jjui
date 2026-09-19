package render

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/idursun/jjui/internal/ui/colorcontrast"
	"github.com/idursun/jjui/internal/ui/layout"
	"github.com/stretchr/testify/assert"
)

func resolveTestRGB(c color.Color) (colorcontrast.RGB, bool) {
	if c == nil {
		return colorcontrast.RGB{}, false
	}
	if _, ok := c.(lipgloss.NoColor); ok {
		return colorcontrast.RGB{}, false
	}
	r, g, b, _ := c.RGBA()
	return colorcontrast.RGB{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}, true
}

func TestForegroundContrastPreservesVisibleBackground(t *testing.T) {
	for _, bg := range []string{"#303030", "#dddddd"} {
		for _, reverse := range []bool{false, true} {
			dl := NewDisplayContext()
			rect := layout.Rect(0, 0, 12, 1)
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(bg)).Background(lipgloss.Color(bg)).Reverse(reverse).Bold(true).Underline(true)
			dl.AddDraw(rect, style.Render("界e\u0301🙂x"), 0)
			before := uv.NewScreenBuffer(12, 1)
			dl.Render(before)
			NewForegroundContrast(resolveTestRGB).Add(dl, rect, 1)
			after := uv.NewScreenBuffer(12, 1)
			dl.Render(after)
			for x := 0; x < 12; x++ {
				original, got := before.CellAt(x, 0), after.CellAt(x, 0)
				assert.Equal(t, original.Content, got.Content)
				assert.Equal(t, original.Width, got.Width)
				assert.Equal(t, original.Style.Attrs, got.Style.Attrs)
				assert.Equal(t, original.Style.Underline, got.Style.Underline)
				if original.Width == 0 || original.Style.Fg == nil {
					continue
				}
				fg, bg := got.Style.Fg, got.Style.Bg
				if reverse {
					assert.Equal(t, original.Style.Fg, got.Style.Fg)
					fg, bg = bg, fg
				} else {
					assert.Equal(t, original.Style.Bg, got.Style.Bg)
				}
				f, _ := resolveTestRGB(fg)
				b, _ := resolveTestRGB(bg)
				assert.GreaterOrEqual(t, colorcontrast.Ratio(f, b), colorcontrast.Minimum)
			}
		}
	}
}

func TestForegroundContrastScopeAndLayers(t *testing.T) {
	dl := NewDisplayContext()
	poor := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Background(lipgloss.Color("#888888"))
	good := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#000000"))
	dl.AddDraw(layout.Rect(0, 0, 4, 1), poor.Render("abcd"), 0)
	dl.AddDraw(layout.Rect(0, 1, 4, 1), good.Render("good"), 0)
	before := uv.NewScreenBuffer(4, 2)
	dl.Render(before)
	contrast := NewForegroundContrast(resolveTestRGB)
	contrast.Add(dl, layout.Rect(-1, 0, 3, 2), 1)
	dl.AddDraw(layout.Rect(0, 0, 1, 1), poor.Render("Z"), 2)
	after := uv.NewScreenBuffer(4, 2)
	dl.Render(after)
	assert.Equal(t, before.CellAt(0, 0).Style, after.CellAt(0, 0).Style, "overlay drawn above effect is untouched")
	assert.NotEqual(t, before.CellAt(1, 0).Style.Fg, after.CellAt(1, 0).Style.Fg)
	assert.Equal(t, before.CellAt(2, 0), after.CellAt(2, 0), "outside selection")
	assert.Equal(t, before.CellAt(0, 1), after.CellAt(0, 1), "readable text")
	assert.Len(t, contrast.cache, 2)
}

func BenchmarkSelectedRowContrast(b *testing.B) {
	for _, enabled := range []bool{false, true} {
		name := "before"
		if enabled {
			name = "after"
		}
		b.Run(name, func(b *testing.B) {
			rect := layout.Rect(0, 0, 100, 1)
			content := lipgloss.NewStyle().Foreground(lipgloss.Color("#b695f3")).Background(lipgloss.Color("#383e56")).Render("selected row with coloured revision IDs, bookmarks and description")
			buf := uv.NewScreenBuffer(100, 1)
			b.ReportAllocs()
			for b.Loop() {
				dl := NewDisplayContext()
				dl.AddDraw(rect, content, 0)
				if enabled {
					NewForegroundContrast(resolveTestRGB).Add(dl, rect, 1)
				}
				dl.Render(buf)
			}
		})
	}
}
