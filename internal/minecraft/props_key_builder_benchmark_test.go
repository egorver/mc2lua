package minecraft

import "testing"

func BenchmarkPropsKeyBuilder_Run(b *testing.B) {
	props := map[string]string{
		"axis":   "x",
		"lit":    "true",
		"half":   "bottom",
		"facing": "north",
		"water":  "false",
	}
	svc := NewPropsKeyBuilder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSink = svc.Run(props)
	}
}
