package pdf2md

import (
	"context"
	"strings"
	"testing"
)

func TestEngineConvert(t *testing.T) {
	engine := NewEngine()
	opts := DefaultOptions()

	res, err := engine.ConvertFile(context.Background(), "../../testdata/sample.pdf", opts)
	if err != nil {
		// Try relative to workspace root if ran from root
		res, err = engine.ConvertFile(context.Background(), "testdata/sample.pdf", opts)
		if err != nil {
			t.Fatalf("conversion failed: %v", err)
		}
	}

	if res.PageCount < 1 {
		t.Errorf("expected at least 1 page, got %d", res.PageCount)
	}

	if len(strings.TrimSpace(res.Markdown)) == 0 {
		t.Fatal("expected non-empty markdown output")
	}

	if !strings.Contains(res.Markdown, "Hello") {
		t.Errorf("expected markdown to contain 'Hello', got:\n%s", res.Markdown)
	}
}
