package validation

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"mini-fhir/internal/fhir/dstu3"
)

func TestProfileCacheVersionInvalidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"resourceType":"StructureDefinition","type":"Patient","snapshot":{"element":[{"path":"Patient.id","min":0}]}}`))
	}))
	defer server.Close()

	tmp := t.TempDir()
	store := NewProfileStore(tmp, time.Hour, 1)
	if _, err := store.loadProfileRules(context.Background(), server.Client(), server.URL); err != nil {
		t.Fatalf("initial load failed: %v", err)
	}

	files, err := os.ReadDir(tmp)
	if err != nil || len(files) == 0 {
		t.Fatalf("expected cache files")
	}
	cachePath := filepath.Join(tmp, files[0].Name())
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("expected cache file")
	}

	store.version = 2
	if cached, ok := store.readCache(server.URL); ok && cached != nil {
		t.Fatalf("expected cache invalidation on version change")
	}
}

func TestProfileCacheTTLInvalidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"resourceType":"StructureDefinition","type":"Patient","snapshot":{"element":[{"path":"Patient.id","min":0}]}}`))
	}))
	defer server.Close()

	tmp := t.TempDir()
	store := NewProfileStore(tmp, time.Millisecond, CacheVersion)
	if _, err := store.loadProfileRules(context.Background(), server.Client(), server.URL); err != nil {
		t.Fatalf("initial load failed: %v", err)
	}

	time.Sleep(5 * time.Millisecond)
	if cached, ok := store.readCache(server.URL); ok && cached != nil {
		t.Fatalf("expected cache invalidation after TTL")
	}
}

func TestProfileStoreLoadDefaults(t *testing.T) {
	registry := dstu3.NewRegistry()
	store := NewProfileStore(t.TempDir(), 0, CacheVersion)
	if err := store.LoadDefaults(context.Background(), registry); err == nil {
		t.Fatalf("expected failure without reachable profile sources")
	}
}
