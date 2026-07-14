package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestEvidenceStoreOpensLocalEvidenceInsideConfiguredDirectory(t *testing.T) {
	root := t.TempDir()
	objectPath := filepath.Join(root, "runs", "run-1", "screenshots", "screen.png")
	if err := os.MkdirAll(filepath.Dir(objectPath), 0o755); err != nil {
		t.Fatalf("create evidence dir: %v", err)
	}
	if err := os.WriteFile(objectPath, []byte("png bytes"), 0o644); err != nil {
		t.Fatalf("write evidence object: %v", err)
	}

	store := NewEvidenceStore(Config{EvidenceDir: root})
	object, err := store.Open(context.Background(), Evidence{
		ID:  "evidence-1",
		URI: "file://" + objectPath,
		Metadata: map[string]any{
			"filename":     "screen.png",
			"content_type": "image/png",
			"size_bytes":   float64(9),
		},
	})
	if err != nil {
		t.Fatalf("open evidence object: %v", err)
	}
	defer object.Body.Close()

	body, err := io.ReadAll(object.Body)
	if err != nil {
		t.Fatalf("read evidence object: %v", err)
	}
	if string(body) != "png bytes" {
		t.Fatalf("unexpected body %q", string(body))
	}
	if object.ContentType != "image/png" {
		t.Fatalf("unexpected content type %q", object.ContentType)
	}
	if object.ContentLength != 9 {
		t.Fatalf("unexpected content length %d", object.ContentLength)
	}
	if object.Filename != "screen.png" {
		t.Fatalf("unexpected filename %q", object.Filename)
	}
}

func TestEvidenceStoreRejectsLocalEvidenceOutsideConfiguredDirectory(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "screen.png")
	if err := os.WriteFile(outside, []byte("png bytes"), 0o644); err != nil {
		t.Fatalf("write outside object: %v", err)
	}

	store := NewEvidenceStore(Config{EvidenceDir: root})
	_, err := store.Open(context.Background(), Evidence{
		ID:       "evidence-1",
		URI:      "file://" + outside,
		Metadata: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected outside evidence path to be rejected")
	}
}
