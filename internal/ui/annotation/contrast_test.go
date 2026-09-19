package annotation

import (
	"math/rand/v2"
	"testing"

	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/idursun/jjui/internal/config"
	"github.com/idursun/jjui/internal/ui/common"
	appContext "github.com/idursun/jjui/internal/ui/context"
	"github.com/idursun/jjui/internal/ui/layout"
	"github.com/idursun/jjui/internal/ui/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContrastRatioReferenceValues(t *testing.T) {
	assert.Equal(t, 21.0, contrastRatio(contrastRGB{}, contrastRGB{255, 255, 255}))
	assert.Equal(t, 1.0, contrastRatio(contrastRGB{127, 127, 127}, contrastRGB{127, 127, 127}))
	// Relative luminance of sRGB red is 0.2126.
	assert.InDelta(t, 5.252, contrastRatio(contrastRGB{255, 0, 0}, contrastRGB{}), 0.00001)
}

func TestScreenshotContrast(t *testing.T) {
	// Sampled solid foreground/background pixels, excluding antialiasing.
	foreground := contrastRGB{182, 149, 243}
	surface := contrastRGB{40, 42, 53}
	for _, tt := range []struct {
		name             string
		background       contrastRGB
		originalContrast float64
	}{
		{"removed word", contrastRGB{134, 63, 67}, 3.059},
		{"added word", contrastRGB{79, 139, 87}, 1.662},
		{"removed line", contrastRGB{83, 50, 58}, 4.553},
		{"added line", contrastRGB{56, 88, 66}, 3.248},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.originalContrast, contrastRatio(foreground, tt.background), 0.001)
			bg := readableBackground(tt.background, surface)
			fg := readableForeground(foreground, bg)
			assert.GreaterOrEqual(t, contrastRatio(fg, bg), syntaxTextContrast)
			// Keep this purple recognisably purple and light on the dark surface.
			assert.GreaterOrEqual(t, fg[0], foreground[0])
			assert.Greater(t, fg[2], fg[0])
			assert.Greater(t, fg[0], fg[1])
			if tt.originalContrast >= syntaxTextContrast {
				assert.Equal(t, foreground, fg)
				assert.Equal(t, tt.background, bg)
			}
		})
	}
}

func TestReadableForegroundAcrossColourPairs(t *testing.T) {
	random := rand.New(rand.NewPCG(1, 2))
	for range 10000 {
		fg := contrastRGB{uint8(random.Uint32()), uint8(random.Uint32()), uint8(random.Uint32())}
		bg := contrastRGB{uint8(random.Uint32()), uint8(random.Uint32()), uint8(random.Uint32())}
		got := readableForeground(fg, bg)
		require.GreaterOrEqual(t, contrastRatio(got, bg), syntaxTextContrast, "fg=%v bg=%v corrected=%v", fg, bg, got)
		if contrastRatio(fg, bg) >= syntaxTextContrast {
			require.Equal(t, fg, got)
		}
	}
}

func TestContrastEffectPreservesGraphemesAndAttributes(t *testing.T) {
	buffer := uv.NewScreenBuffer(12, 1)
	rect := layout.Rect(0, 0, 12, 1)
	text := lipgloss.NewStyle().Foreground(lipgloss.Color("#b695f3")).Background(lipgloss.Color("#4f8b57")).Bold(true).Italic(true).Underline(true).Render("界e\u0301🙂x")
	uv.NewStyledString(text).Draw(buffer, rect)
	before := make([]*uv.Cell, 12)
	for x := range 12 {
		before[x] = buffer.CellAt(x, 0).Clone()
	}
	adjuster := newContrastAdjuster(nil, "#282a35", nil, true)
	effect := annotationContrastEffect{rect: rect, adjuster: adjuster}
	effect.Apply(buffer)
	for x := range 12 {
		got := buffer.CellAt(x, 0)
		assert.Equal(t, before[x].Content, got.Content, "x=%d", x)
		assert.Equal(t, before[x].Width, got.Width, "x=%d", x)
		assert.Equal(t, before[x].Style.Attrs, got.Style.Attrs)
		assert.Equal(t, before[x].Style.Underline, got.Style.Underline)
		if got.Width > 0 && got.Style.Fg != nil && got.Style.Bg != nil {
			assert.GreaterOrEqual(t, contrastRatio(adjuster.rgb(got.Style.Fg), adjuster.rgb(got.Style.Bg)), syntaxTextContrast)
		}
	}
	assert.Len(t, adjuster.pairs, 1, "one calculation per colour pair, not per cell")
}

func TestContrastEffectClipsAndLeavesUnknownBackgroundAlone(t *testing.T) {
	buffer := uv.NewScreenBuffer(3, 1)
	uv.NewStyledString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("abc")).Draw(buffer, layout.Rect(0, 0, 3, 1))
	before := buffer.CellAt(1, 0).Clone()
	effect := annotationContrastEffect{rect: layout.Rect(-2, -2, 10, 10), adjuster: newContrastAdjuster(nil, "", nil, true)}
	effect.Apply(buffer)
	assert.Equal(t, before, buffer.CellAt(1, 0))
}

func TestContrastEffectUsesReportedANSIColours(t *testing.T) {
	adjuster := newContrastAdjuster(nil, "#282a35", map[int]string{5: "#b695f3", 2: "#4f8b57"}, true)
	pair := contrastPair{ansi.Magenta, ansi.Green}
	got := adjuster.adjust(pair)
	assert.GreaterOrEqual(t, contrastRatio(adjuster.rgb(got.foreground), adjuster.rgb(got.background)), syntaxTextContrast)
	assert.NotEqual(t, pair, got)
	// A theme/palette reload starts a new frame with no old resolved pairs.
	reloaded := newContrastAdjuster(nil, "#282a35", map[int]string{5: "#ffffff", 2: "#000000"}, true)
	assert.Equal(t, pair, reloaded.adjust(pair))
}

func TestAnnotationContrastAfterDiffAndSelectionComposition(t *testing.T) {
	previous := common.DefaultPalette
	t.Cleanup(func() { common.DefaultPalette = previous })
	for _, tt := range []struct {
		name, surface, text, syntax, added, deleted, selected string
		dark                                                  bool
	}{
		{"screenshot", "#282a35", "#f8f8f3", "#b695f3", "#88ef8d", "#ef6d67", "#383e56", true},
		{"light", "#ffffff", "#24292f", "#8250df", "#228822", "#cc3333", "#c0c0c0", false},
		{"dark saturated", "#000000", "#ffffff", "#000080", "#00ff00", "#ff0000", "#808080", true},
		{"light pale syntax", "#fffaf0", "#222222", "#aaddaa", "#66dd66", "#ff8888", "#dddddd", false},
		{"midgrey", "#777777", "#ffffff", "#888888", "#88aa88", "#aa8888", "#999999", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			common.DefaultPalette = common.NewPalette()
			common.DefaultPalette.Update(map[string]config.Color{
				"syntax text": {Fg: tt.text}, "syntax keyword": {Fg: tt.syntax},
				"syntax string": {Fg: tt.syntax}, "syntax function": {Fg: tt.syntax},
				"syntax comment": {Fg: tt.syntax}, "syntax type": {Fg: tt.syntax},
				"added": {Fg: tt.added}, "deleted": {Fg: tt.deleted},
				"annotation:selected": {Bg: tt.selected},
			})
			common.DefaultPalette.ConfigureBackgroundBlend(0.4, tt.surface, nil)
			model := mouseTestModel(4)
			model.context = &appContext.MainContext{TerminalBackground: tt.surface, TerminalHasDarkBackground: tt.dark}
			lines := model.document.files[0].Patch.Lines
			lines[0] = patchLine{Kind: lineRemoved, OldLine: 1, Content: `return "old"`, PairContent: `return "new"`}
			lines[1] = patchLine{Kind: lineAdded, NewLine: 1, Content: `return "new"`, PairContent: `return "old"`}
			lines[2].Content = `return "unchanged"`
			lines[3].Content = `return "selected"`
			model.cursor = 3
			buffer := uv.NewScreenBuffer(50, 6)
			display := render.NewDisplayContext()
			model.ViewRect(display, layout.NewBox(layout.Rect(0, 0, 50, 6)))
			display.Render(buffer)
			resolver := newContrastAdjuster(nil, tt.surface, nil, tt.dark)
			for _, y := range []int{1, 2, 4} {
				checked := 0
				for x := contentColumn; x < 40; x++ {
					cell := buffer.CellAt(x, y)
					if cell.Content == " " || cell.Style.Fg == nil || cell.Style.Bg == nil {
						continue
					}
					checked++
					assert.GreaterOrEqual(t, contrastRatio(resolver.rgb(cell.Style.Fg), resolver.rgb(cell.Style.Bg)), syntaxTextContrast, "x=%d y=%d", x, y)
				}
				assert.Positive(t, checked)
			}
			assert.Nil(t, buffer.CellAt(contentColumn, 3).Style.Bg, "unhighlighted context should retain its styling")
			assert.NotEqual(t, buffer.CellAt(contentColumn, 1).Style.Bg, buffer.CellAt(contentColumn+8, 1).Style.Bg, "word and line backgrounds remain distinct")
			assert.NotEqual(t, buffer.CellAt(contentColumn, 2).Style.Bg, buffer.CellAt(contentColumn+8, 2).Style.Bg, "word and line backgrounds remain distinct")
		})
	}
}

func TestContrastBackgroundUsesSurfacePolarity(t *testing.T) {
	for _, tt := range []struct {
		name                      string
		surface, background, edge contrastRGB
	}{
		{"dark", contrastRGB{40, 42, 53}, contrastRGB{79, 139, 87}, contrastRGB{255, 255, 255}},
		{"light", contrastRGB{250, 250, 250}, contrastRGB{80, 130, 80}, contrastRGB{0, 0, 0}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			adjusted := readableBackground(tt.background, tt.surface)
			assert.GreaterOrEqual(t, contrastRatio(adjusted, tt.edge), highlightTextContrast)
			// One uniform background change for the entire highlight, even for spaces.
			adjuster := &contrastAdjuster{surface: tt.surface, pairs: make(map[contrastPair]contrastPair)}
			blank := adjuster.adjust(contrastPair{background: tt.background.color()})
			text := adjuster.adjust(contrastPair{foreground: contrastRGB{120, 100, 180}.color(), background: tt.background.color()})
			assert.Equal(t, blank.background, text.background)
			assert.Nil(t, blank.foreground)
		})
	}
	// Respect an explicit annotation surface, even when the terminal uses the
	// opposite mode; transparent/default surfaces follow the terminal instead.
	explicit := newContrastAdjuster(lipgloss.Color("#fafafa"), "#282a35", nil, true)
	assert.Equal(t, contrastRGB{250, 250, 250}, explicit.surface)
	transparent := newContrastAdjuster(lipgloss.NoColor{}, "#282a35", nil, false)
	assert.Equal(t, contrastRGB{40, 42, 53}, transparent.surface)
}

func TestContrastEffectHandlesReverse(t *testing.T) {
	buffer := uv.NewScreenBuffer(1, 1)
	buffer.SetCell(0, 0, &uv.Cell{Content: "x", Width: 1, Style: uv.Style{
		Fg:    ansi.RGBColor{R: 79, G: 139, B: 87},
		Bg:    ansi.RGBColor{R: 182, G: 149, B: 243},
		Attrs: uv.AttrReverse | uv.AttrBold,
	}})
	adjuster := newContrastAdjuster(nil, "#282a35", nil, true)
	annotationContrastEffect{rect: layout.Rect(0, 0, 1, 1), adjuster: adjuster}.Apply(buffer)
	cell := buffer.CellAt(0, 0)
	assert.Equal(t, uint8(uv.AttrReverse|uv.AttrBold), cell.Style.Attrs)
	assert.GreaterOrEqual(t, contrastRatio(adjuster.rgb(cell.Style.Bg), adjuster.rgb(cell.Style.Fg)), syntaxTextContrast)
}
