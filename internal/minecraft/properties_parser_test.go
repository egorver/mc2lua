package minecraft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseProperties(t *testing.T) {
	svc := NewPropertiesParser()
	tests := []struct {
		name string
		json string
		want map[string]string
	}{
		{name: "empty", json: "", want: map[string]string{}},
		{name: "simple", json: `{"a":"1","b":"2"}`, want: map[string]string{"a": "1", "b": "2"}},
		{name: "numeric value", json: `{"x":42}`, want: map[string]string{"x": "42"}},
		{name: "boolean value", json: `{"up":true}`, want: map[string]string{"up": "true"}},
		{name: "invalid json", json: `{bad`, want: map[string]string{}},
		{name: "nested object", json: `{"obj":{"k":"v"}}`, want: map[string]string{"obj": "map[k:v]"}},
		{name: "array value", json: `{"arr":[1,2]}`, want: map[string]string{"arr": "[1 2]"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.Run(tt.json)
			require.Equal(t, tt.want, got)
		})
	}
}
