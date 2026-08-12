package minecraft

import (
	"encoding/json"
	"testing"
)

func FuzzBlockstateParser_ParseVariantValue(f *testing.F) {
	f.Add([]byte(`{"model":"block/cube"}`))
	f.Add([]byte(`[{"model":"block/stone"},{"model":"block/andesite"}]`))
	f.Add([]byte(`{broken`))
	f.Add([]byte(``))
	svc := NewBlockstateParser(nil, NewPropsKeyBuilder())
	f.Fuzz(func(t *testing.T, raw []byte) {
		_, _ = svc.parseVariantValue(json.RawMessage(raw))
	})
}

func FuzzBlockstateParser_MatchMultipart(f *testing.F) {
	f.Add([]byte(`{"apply":{"model":"block/a"}}`))
	f.Add([]byte(`{"when":{"north":"true"},"apply":{"model":"block/side","y":90}}`))
	f.Add([]byte(`{broken`))
	svc := NewBlockstateParser(nil, NewPropsKeyBuilder())
	f.Fuzz(func(t *testing.T, raw []byte) {
		_, _ = svc.matchMultipart([]json.RawMessage{json.RawMessage(raw)}, map[string]string{"north": "true"})
	})
}
