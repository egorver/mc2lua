package minecraft

import (
	"encoding/json"
	"testing"
)

func FuzzTextureResolver_Run(f *testing.F) {
	f.Add([]byte(`{"a":"block/stone"}`))
	f.Add([]byte(`{"a":"#b","b":"block/stone"}`))
	f.Add([]byte(`{"a":"#b","b":"#a"}`))
	f.Add([]byte(`{"a":"#a"}`))
	f.Add([]byte(`{"a":"#missing"}`))
	svc := NewTextureResolver()
	f.Fuzz(func(t *testing.T, raw []byte) {
		var m map[string]string
		if err := json.Unmarshal(raw, &m); err != nil {
			return
		}
		_ = svc.Run(m)
	})
}
