package test

import (
	"testing"

	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/idursun/jjui/internal/ui/colorcontrast"
	"github.com/stretchr/testify/assert"
)

// AssertSelectedContrast locates unique test marker characters in a rendered
// screen. It checks the final cells, not the styles supplied to the renderer.
func AssertSelectedContrast(t *testing.T, buf uv.Screen, selected, unselected, background string) {
	t.Helper()
	selectedCount, unselectedCount := 0, 0
	for y := buf.Bounds().Min.Y; y < buf.Bounds().Max.Y; y++ {
		for x := buf.Bounds().Min.X; x < buf.Bounds().Max.X; x++ {
			c := buf.CellAt(x, y)
			if c == nil {
				continue
			}
			switch c.Content {
			case selected:
				selectedCount++
				fg, bgColor := c.Style.Fg, c.Style.Bg
				if c.Style.Attrs&uv.AttrReverse != 0 {
					fg, bgColor = bgColor, fg
				}
				assert.Equal(t, lipgloss.Color(background), bgColor)
				if !assert.NotNil(t, fg) {
					continue
				}
				r, g, b, _ := fg.RGBA()
				br, bg, bb, _ := bgColor.RGBA()
				assert.GreaterOrEqual(t, colorcontrast.Ratio(colorcontrast.RGB{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}, colorcontrast.RGB{uint8(br >> 8), uint8(bg >> 8), uint8(bb >> 8)}), colorcontrast.Minimum)
			case unselected:
				unselectedCount++
				assert.Equal(t, lipgloss.Color(background), c.Style.Fg, "unselected foreground stays unchanged")
			}
		}
	}
	assert.Positive(t, selectedCount)
	assert.Positive(t, unselectedCount)
}
