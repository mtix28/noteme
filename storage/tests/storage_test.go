package storage_test

import (
	"os"
	"testing"
	"time"

	"github.com/mtix28/noteme/model"
	"github.com/mtix28/noteme/storage"
)

func setupTestStorage(t *testing.T) (*storage.Storage, string) {
	tempDir, err := os.MkdirTemp("", "noteme_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// We need to hack the storage to use our temp dir
	// Since NewStorage uses UserHomeDir, we can't easily mock it without dependency injection.
	// However, for this test, we can just construct the struct manually if the field was exported.
    // It is not exported.
    // Let's create a constructor in storage/storage.go that accepts a path for testing purposes,
    // or just modify the environment variable HOME if running on Linux/Mac, but that's flaky.
    
    // Better approach: Let's modify storage.go to allow setting the base path or move the logic to a method we can override?
    // No, let's just make a NewStorageWithDir for testing or export the struct fields?
    // Actually, NewStorage uses .noteme in Home.
    
    // Simplest for now: set HOME env var to tempDir
    os.Setenv("HOME", tempDir)
    
	store, err := storage.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	return store, tempDir
}

func cleanupTestStorage(path string) {
	os.RemoveAll(path)
}

func TestDeleteNote(t *testing.T) {
	store, dir := setupTestStorage(t)
	defer cleanupTestStorage(dir)

	// Create a note
	note := model.Note{
		ID:        "1",
		Title:     "Test Note",
		Content:   "Content",
		CreatedAt: time.Now(),
		Folder:    "general",
	}
	err := store.SaveNotes([]model.Note{note})
	if err != nil {
		t.Fatalf("Failed to save note: %v", err)
	}

	// Verify saved
	notes, _ := store.LoadNotes()
	if len(notes) != 1 {
		t.Fatalf("Expected 1 note, got %d", len(notes))
	}

	// Delete
	err = store.DeleteNote("1")
	if err != nil {
		t.Fatalf("Failed to delete note: %v", err)
	}

	// Verify deleted
	notes, _ = store.LoadNotes()
	if len(notes) != 0 {
		t.Fatalf("Expected 0 notes, got %d", len(notes))
	}
}

func TestDeleteTodo(t *testing.T) {
	store, dir := setupTestStorage(t)
	defer cleanupTestStorage(dir)

	todo := model.Todo{
		ID:        "t1",
		Content:   "Test Todo",
		Done:      false,
		CreatedAt: time.Now(),
		Frequency: model.Once,
	}

	err := store.SaveTodos([]model.Todo{todo})
	if err != nil {
		t.Fatalf("Failed to save todo: %v", err)
	}

	// Verify saved
	todos, _ := store.LoadTodos()
	if len(todos) != 1 {
		t.Fatalf("Expected 1 todo, got %d", len(todos))
	}

	// Delete
	err = store.DeleteTodo("t1")
	if err != nil {
		t.Fatalf("Failed to delete todo: %v", err)
	}

	// Verify deleted
	todos, _ = store.LoadTodos()
	if len(todos) != 0 {
		t.Fatalf("Expected 0 todos, got %d", len(todos))
	}
}
