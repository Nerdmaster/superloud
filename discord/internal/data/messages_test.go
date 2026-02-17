package data

import (
	"os"
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) (*SQLiteStore, func()) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.Remove(dbPath)
	}

	return store, cleanup
}

func TestMessages_LoadAndRandom(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	// Insert some test messages
	store.InsertMessage(&Message{Text: "FIRST MESSAGE", Author: "Somebody", Score: 1, Views: 0})
	store.InsertMessage(&Message{Text: "SECOND MESSAGE", Author: "Another", Score: 1, Views: 0})

	msgs := NewMessages(store)
	if err := msgs.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Random should return messages
	msg := msgs.Random()
	if msg == nil {
		t.Fatal("Random returned nil")
	}
	if msg.Text != "FIRST MESSAGE" && msg.Text != "SECOND MESSAGE" {
		t.Errorf("unexpected message: %s", msg.Text)
	}
}

func TestMessages_Add(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	msgs := NewMessages(store)
	if err := msgs.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	msgs.Add("NEW LOUD MESSAGE", "TestUser")

	// Should be able to get the message as random
	// First exhaust the default message
	msgs.Random()
	msg := msgs.Random()
	if msg == nil {
		t.Fatal("Random returned nil after add")
	}
}

func TestMessages_VoteDuplicate(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.InsertMessage(&Message{Text: "VOTE TEST", Author: "Someone", Score: 1, Views: 0})

	msgs := NewMessages(store)
	if err := msgs.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	msgs.Random() // Set last message

	userHash := int64(12345)

	// First vote should succeed
	if !msgs.Vote(userHash, 1) {
		t.Error("first vote should succeed")
	}

	// Second vote from same user should fail
	if msgs.Vote(userHash, 1) {
		t.Error("duplicate vote should fail")
	}

	// Different user should succeed
	if !msgs.Vote(userHash+1, -1) {
		t.Error("different user vote should succeed")
	}
}

func TestMessages_Serialize(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	msgs := NewMessages(store)
	if err := msgs.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	msgs.Add("SERIALIZE TEST", "Tester")
	msgs.Serialize()

	// Verify it was written to DB
	exists, err := store.Exists("SERIALIZE TEST")
	if err != nil {
		t.Fatalf("Exists check failed: %v", err)
	}
	if !exists {
		t.Error("message should exist in DB after serialize")
	}
}

func TestMessages_VoteRandomClearsVoters(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	store.InsertMessage(&Message{Text: "MSG ONE", Author: "A", Score: 1, Views: 0})
	store.InsertMessage(&Message{Text: "MSG TWO", Author: "B", Score: 1, Views: 0})

	msgs := NewMessages(store)
	if err := msgs.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	msgs.Random()
	userHash := int64(99999)
	msgs.Vote(userHash, 1)

	// Getting a new random message should clear voters
	msgs.Random()
	if !msgs.Vote(userHash, 1) {
		t.Error("vote should succeed after new random message")
	}
}
