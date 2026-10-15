package pdf2md

import (
	"context"
	"testing"
)

func BenchmarkEngineConvert(b *testing.B) {
	engine := NewEngine()
	opts := DefaultOptions()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.ConvertFile(ctx, "../../testdata/sample.pdf", opts)
		if err != nil {
			_, err = engine.ConvertFile(ctx, "testdata/sample.pdf", opts)
			if err != nil {
				b.Fatalf("benchmark iteration failed: %v", err)
			}
		}
	}
}
