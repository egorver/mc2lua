package minecraft

import "testing"

func BenchmarkModelParser_Flatten(b *testing.B) {
	fs, nsToRoots := setupTestFS()
	addModel(fs, "minecraft", "block/cube", []byte(testModelCube))
	addModel(fs, "minecraft", "block/grass", []byte(testModelGrass))
	addModel(fs, "minecraft", "block/grass_block", []byte(testModelGrandchild))

	svc := NewModelParser(fs)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fm, err := svc.Run("minecraft:block/grass_block", nsToRoots)
		if err != nil {
			b.Fatal(err)
		}
		benchSink = fm
	}
}
