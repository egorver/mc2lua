package pipeline

import (
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

func TestCuboidHelperRoundTrip(t *testing.T) {
	t.Parallel()

	helper := NewCuboidHelper()
	values := []int{-5, -2, -1, 0, 1, 3, 100}
	sizes := []int{1, 2, 3, 4, 5}

	for _, x := range values {
		for _, w := range sizes {
			gotX, _, _ := helper.MinCorner(model.Cuboid{X: helper.Center(x, w), Width: w})
			require.Equal(t, x, gotX)
		}
	}
	for _, y := range values {
		for _, h := range sizes {
			_, gotY, _ := helper.MinCorner(model.Cuboid{Y: helper.Center(y, h), Height: h})
			require.Equal(t, y, gotY)
		}
	}
	for _, z := range values {
		for _, d := range sizes {
			_, _, gotZ := helper.MinCorner(model.Cuboid{Z: helper.Center(z, d), Depth: d})
			require.Equal(t, z, gotZ)
		}
	}
}
