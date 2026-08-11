package main

import (
	"bytes"
	"errors"
	"math"
	"strings"
	"testing"

	"mc2lua/internal/app"

	"github.com/stretchr/testify/require"
)

type mockRunner struct {
	runFn func(cfg app.AppConfig) error
	got   app.AppConfig
}

func (m *mockRunner) Run(cfg app.AppConfig) error {
	m.got = cfg
	if m.runFn != nil {
		return m.runFn(cfg)
	}
	return nil
}

func TestRun_FlagsMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want app.AppConfig
	}{
		{
			name: "defaults",
			args: nil,
			want: app.AppConfig{
				Input:         defaultInputDir,
				AssetsDir:     defaultAssetsDir,
				Output:        defaultOutputFile,
				PartsTemplate: defaultPartsTemplate,
				Scale:         defaultScale,
				ConfigDir:     defaultConfigDir,
				XMin:          math.MinInt32,
				XMax:          math.MaxInt32,
				YMin:          math.MinInt32,
				YMax:          math.MaxInt32,
				ZMin:          math.MinInt32,
				ZMax:          math.MaxInt32,
			},
		},
		{
			name: "all flags set",
			args: []string{
				"-input", "reg",
				"-assets", "a",
				"-output", "o.lua",
				"-parts-template", "t.yaml",
				"-scale", "2",
				"-no-offset",
				"-config", "cfg",
				"-xmin", "-100", "-xmax", "100",
				"-ymin", "-50", "-ymax", "50",
				"-zmin", "-25", "-zmax", "25",
			},
			want: app.AppConfig{
				Input:         "reg",
				AssetsDir:     "a",
				Output:        "o.lua",
				PartsTemplate: "t.yaml",
				Scale:         2,
				NoOffset:      true,
				ConfigDir:     "cfg",
				XMin:          -100,
				XMax:          100,
				YMin:          -50,
				YMax:          50,
				ZMin:          -25,
				ZMax:          25,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			runner := &mockRunner{}
			code := run(tt.args, &stdout, &stderr, func() appRunner { return runner })
			require.Zero(t, code)
			require.Equal(t, tt.want, runner.got)
			require.Empty(t, stderr.String())
		})
	}
}

func TestRun_Help(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{{"-h"}, {"--help"}, {"-help"}} {
		args := args
		t.Run(strings.Join(args, ""), func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			code := run(args, &stdout, &stderr, func() appRunner { return &mockRunner{} })
			require.Zero(t, code)
			require.Contains(t, stderr.String(), "Usage of")
		})
	}
}

func TestRun_UnknownFlag(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	code := run([]string{"-nope"}, &stdout, &stderr, func() appRunner { return &mockRunner{} })
	require.NotZero(t, code)
	require.Contains(t, stderr.String(), "flag provided but not defined")
}

func TestRun_RunnerError(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	runner := &mockRunner{runFn: func(cfg app.AppConfig) error {
		return errors.New("boom")
	}}
	code := run(nil, &stdout, &stderr, func() appRunner { return runner })
	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "Error: boom")
}
