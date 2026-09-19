package render

import (
	"image/color"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/idursun/jjui/internal/ui/colorcontrast"
	"github.com/idursun/jjui/internal/ui/layout"
)

type foregroundPair struct{ foreground, background ansi.Color }

// ForegroundContrast corrects visible foregrounds without changing visible
// backgrounds. Create one per frame to avoid stale terminal/theme colours.
type ForegroundContrast struct {
	resolve func(color.Color) (colorcontrast.RGB, bool)
	cache   map[foregroundPair]ansi.Color
}

func NewForegroundContrast(resolve func(color.Color) (colorcontrast.RGB, bool)) *ForegroundContrast {
	return &ForegroundContrast{resolve: resolve, cache: make(map[foregroundPair]ansi.Color)}
}

// Add registers correction after the selected row's draws and background
// effects, but before any higher-layer dialogs. Only this rectangle is changed.
func (c *ForegroundContrast) Add(dl *DisplayContext, rect layout.Rectangle, z int) {
	dl.AddEffect(foregroundContrastEffect{rect: rect, z: z, contrast: c})
}

func (c *ForegroundContrast) foreground(pair foregroundPair) ansi.Color {
	if fg, ok := c.cache[pair]; ok {
		return fg
	}
	result := pair.foreground
	fg, fgOK := c.resolve(pair.foreground)
	bg, bgOK := c.resolve(pair.background)
	if fgOK && bgOK {
		adjusted := colorcontrast.Foreground(fg, bg)
		if adjusted != fg {
			result = adjusted.Color()
		}
	}
	c.cache[pair] = result
	return result
}

type foregroundContrastEffect struct {
	rect     layout.Rectangle
	z        int
	contrast *ForegroundContrast
}

func (e foregroundContrastEffect) GetRect() layout.Rectangle { return e.rect }
func (e foregroundContrastEffect) GetZ() int                 { return e.z }
func (e foregroundContrastEffect) Apply(buf uv.Screen) {
	iterateCells(buf, e.rect, func(cell *uv.Cell) *uv.Cell {
		if cell == nil {
			return nil
		}
		pair := foregroundPair{cell.Style.Fg, cell.Style.Bg}
		reverse := cell.Style.Attrs&uv.AttrReverse != 0
		if reverse {
			pair.foreground, pair.background = pair.background, pair.foreground
		}
		fg := e.contrast.foreground(pair)
		if fg == pair.foreground {
			return nil
		}
		updated := cell.Clone()
		// With reverse video the stored background is the visible foreground.
		if reverse {
			updated.Style.Bg = fg
		} else {
			updated.Style.Fg = fg
		}
		return updated
	})
}
