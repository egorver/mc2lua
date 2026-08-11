package pipeline

import (
	"errors"
	"strings"
	"testing"

	"mc2lua/internal/model"
	"mc2lua/internal/runtime"

	"github.com/stretchr/testify/require"
)

func TestTemplateGenerator_New(t *testing.T) {
	t.Parallel()

	tg := NewTemplateGenerator(runtime.NewFSMock(), &mockPartStyleMatcher{})
	require.NotNil(t, tg)
}

func TestTemplateGenerator_Run(t *testing.T) {
	t.Parallel()

	parts := []model.Part{
		{BlockID: "minecraft:stone", Color: model.Color{100, 100, 100}},
		{BlockID: "minecraft:stone", Color: model.Color{120, 120, 120}},
		{BlockID: "minecraft:deepslate", Color: model.Color{70, 72, 74}},
		{BlockID: "minecraft:oak_leaves", Color: model.Color{61, 103, 45}},
		{BlockID: "minecraft:stone_config", Color: model.Color{10, 20, 30}},
	}
	blocks := []model.RawBlock{
		{ID: "minecraft:stone"},
		{ID: "minecraft:stone"},
		{ID: "minecraft:stone"},
		{ID: "minecraft:stone"},
		{ID: "minecraft:stone"},
		{ID: "minecraft:deepslate"},
		{ID: "minecraft:deepslate"},
		{ID: "minecraft:oak_leaves"},
		{ID: "minecraft:stone_config"},
	}
	configured := model.Color{1, 2, 3}
	styles := map[string]model.PartStyle{
		"minecraft:stone_config": {Color: &configured},
	}

	svc := NewTemplateGenerator(runtime.NewFSMock(), &mockPartStyleMatcher{styles: styles})
	mockFS := runtime.NewFSMock()
	svc.fs = mockFS

	err := svc.Run(parts, blocks, "out.yaml")
	require.NoError(t, err)

	data, err := mockFS.ReadFile("out.yaml")
	require.NoError(t, err)
	got := string(data)

	require.Contains(t, got, "minecraft:stone:")
	require.Contains(t, got, "minecraft:deepslate:")
	require.Contains(t, got, "minecraft:oak_leaves:")
	require.NotContains(t, got, "minecraft:stone_config")

	require.Contains(t, got, "color: [110, 110, 110]")
	require.Contains(t, got, "color: [70, 72, 74]")
	require.Contains(t, got, "color: [61, 103, 45]")

	stoneIdx := strings.Index(got, "minecraft:stone:")
	deepslateIdx := strings.Index(got, "minecraft:deepslate:")
	oakIdx := strings.Index(got, "minecraft:oak_leaves:")
	require.True(t, stoneIdx < deepslateIdx && deepslateIdx < oakIdx,
		"blocks should be ordered by popularity, got:\n%s", got)
}

func TestTemplateGenerator_Run_OrdersEqualPopularityById(t *testing.T) {
	t.Parallel()

	parts := []model.Part{
		{BlockID: "minecraft:b", Color: model.Color{1, 1, 1}},
		{BlockID: "minecraft:a", Color: model.Color{2, 2, 2}},
	}
	blocks := []model.RawBlock{
		{ID: "minecraft:b"},
		{ID: "minecraft:a"},
	}

	svc := NewTemplateGenerator(runtime.NewFSMock(), &mockPartStyleMatcher{})
	mockFS := runtime.NewFSMock()
	svc.fs = mockFS

	err := svc.Run(parts, blocks, "out.yaml")
	require.NoError(t, err)

	data, err := mockFS.ReadFile("out.yaml")
	require.NoError(t, err)
	got := string(data)

	aIdx := strings.Index(got, "minecraft:a:")
	bIdx := strings.Index(got, "minecraft:b:")
	require.True(t, aIdx < bIdx, "equal popularity should be ordered by id, got:\n%s", got)
}

func TestTemplateGenerator_Run_Empty(t *testing.T) {
	t.Parallel()

	svc := NewTemplateGenerator(runtime.NewFSMock(), &mockPartStyleMatcher{})
	mockFS := runtime.NewFSMock()
	svc.fs = mockFS

	err := svc.Run(nil, nil, "out.yaml")
	require.NoError(t, err)

	data, err := mockFS.ReadFile("out.yaml")
	require.NoError(t, err)
	require.Equal(t, "parts: {}\n", string(data))
}

func TestTemplateGenerator_Run_AllConfigured(t *testing.T) {
	t.Parallel()

	configured := model.Color{1, 2, 3}
	styles := map[string]model.PartStyle{
		"minecraft:stone": {Color: &configured},
	}

	svc := NewTemplateGenerator(runtime.NewFSMock(), &mockPartStyleMatcher{styles: styles})
	mockFS := runtime.NewFSMock()
	svc.fs = mockFS

	parts := []model.Part{{BlockID: "minecraft:stone", Color: model.Color{10, 10, 10}}}
	blocks := []model.RawBlock{{ID: "minecraft:stone"}}

	err := svc.Run(parts, blocks, "out.yaml")
	require.NoError(t, err)

	data, err := mockFS.ReadFile("out.yaml")
	require.NoError(t, err)
	require.Equal(t, "parts: {}\n", string(data))
}

func TestTemplateGenerator_Run_WriteError(t *testing.T) {
	t.Parallel()

	svc := NewTemplateGenerator(runtime.NewFSMock(), &mockPartStyleMatcher{})
	mockFS := runtime.NewFSMock()
	mockFS.CreateErrors["out.yaml"] = errors.New("create fail")
	svc.fs = mockFS

	err := svc.Run(nil, nil, "out.yaml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write parts template file")
}
