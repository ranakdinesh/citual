package app

import (
	"testing"

	"github.com/google/uuid"
)

func TestStorageObjectIDFromURLAcceptsAppRelativeStorageObject(t *testing.T) {
	id := uuid.New()

	got, ok := storageObjectIDFromURL("/storage/objects/" + id.String())
	if !ok {
		t.Fatal("expected storage object URL to be accepted")
	}
	if got != id {
		t.Fatalf("expected %s, got %s", id, got)
	}
}

func TestStorageObjectIDFromURLRejectsNonStorageURL(t *testing.T) {
	if _, ok := storageObjectIDFromURL("https://example.com/logo.png"); ok {
		t.Fatal("expected external URL to be rejected")
	}
	if _, ok := storageObjectIDFromURL("/storage/objects/not-a-uuid"); ok {
		t.Fatal("expected malformed storage object URL to be rejected")
	}
}
