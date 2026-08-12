package pipeline

import (
	"bytes"
	"testing"
	"time"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

func TestStress_LuaGenerator_LargeOutput(t *testing.T) {
	const total = 100_000

	parts := make([]model.Part, 0, total)
	for i := 0; i < total; i++ {
		groupID := 0
		group := ""
		if i%2 == 1 {
			groupID = 1
			group = "Group1"
		}
		parts = append(parts, makePart(
			"stone", group, groupID, "minecraft:stone", "",
			model.Vector3{1, 1, 1}, model.Vector3{1, 1, 1},
			model.Color{125, 125, 125}, "Slate",
		))
	}

	svc, mockFS := testLuaGenerator(t)
	start := time.Now()
	require.NoError(t, svc.Run(parts, 4, "out.lua"))
	t.Logf("generated lua for %d parts in %s", total, time.Since(start))

	data, err := mockFS.ReadFile("out.lua")
	require.NoError(t, err)
	require.True(t, bytes.Contains(data, []byte("-- Total parts: 100000")))
	require.True(t, bytes.Contains(data, []byte("local _total = 100000")))
	require.True(t, bytes.Contains(data, []byte(`_group.Name = "Group1"`)))
}
