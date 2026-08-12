package minecraft

import (
	"fmt"
	"testing"

	"mc2lua/internal/runtime"
)

func BenchmarkAssetScanner_Run(b *testing.B) {
	fs := runtime.NewFSMock()
	fs.AddDir("assets", 0755)
	fs.AddDir("assets/mod", 0755)
	for i := 0; i < 100; i++ {
		ns := fmt.Sprintf("namespace_%d", i)
		fs.AddDir("assets/mod/"+ns, 0755)
		fs.AddDir("assets/mod/"+ns+"/blockstates", 0755)
	}
	scanner := NewAssetScanner(fs)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := scanner.Run("assets")
		if err != nil {
			b.Fatal(err)
		}
		benchSink = result
	}
}
