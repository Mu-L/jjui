package colorcontrast

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContrastRatioReferenceValues(t *testing.T) {
	assert.Equal(t, 21.0, Ratio(RGB{}, RGB{255, 255, 255}))
	assert.Equal(t, 1.0, Ratio(RGB{127, 127, 127}, RGB{127, 127, 127}))
	// Relative luminance of sRGB red is 0.2126.
	assert.InDelta(t, 5.252, Ratio(RGB{255, 0, 0}, RGB{}), 0.00001)
}

func TestReadableForegroundAcrossColourPairs(t *testing.T) {
	random := rand.New(rand.NewPCG(1, 2))
	for range 10000 {
		fg := RGB{uint8(random.Uint32()), uint8(random.Uint32()), uint8(random.Uint32())}
		bg := RGB{uint8(random.Uint32()), uint8(random.Uint32()), uint8(random.Uint32())}
		got := Foreground(fg, bg)
		require.GreaterOrEqual(t, Ratio(got, bg), Minimum, "fg=%v bg=%v corrected=%v", fg, bg, got)
		if Ratio(fg, bg) >= Minimum {
			require.Equal(t, fg, got)
		}
	}
}
