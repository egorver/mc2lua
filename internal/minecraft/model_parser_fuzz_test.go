package minecraft

import (
	"encoding/json"
	"testing"
)

func FuzzModelParser_ParseElements(f *testing.F) {
	f.Add([]byte(`{"from":[0,0,0],"to":[16,16,16]}`))
	f.Add([]byte(`{"from":[0,0,0],"to":[16,16,16],"faces":{"north":{"uv":[0,0,16,16],"texture":"#all"}}}`))
	f.Add([]byte(`{bad`))
	svc := &ModelParser{}
	f.Fuzz(func(t *testing.T, raw []byte) {
		_, _ = svc.parseElements([]json.RawMessage{json.RawMessage(raw)})
	})
}
