package minecraft

import (
	"testing"
)

func FuzzPropertiesParser_Run(f *testing.F) {
	f.Add(`{"facing":"north"}`)
	f.Add(`{"a":1,"b":true,"c":[1,2]}`)
	f.Add(`{bad`)
	f.Add(``)
	svc := NewPropertiesParser()
	f.Fuzz(func(t *testing.T, jsonStr string) {
		_ = svc.Run(jsonStr)
	})
}
