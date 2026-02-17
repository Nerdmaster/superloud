package data

import (
	"log"
	"math/rand"
	"sync"
)

// Messages is the in-memory message container with weighted random selection,
// voting, and dirty tracking for serialization.
type Messages struct {
	mu sync.Mutex

	store    *SQLiteStore
	messages map[string]*Message // keyed by text
	random   []string            // shuffled queue of message texts
	Last     *Message

	dirty           bool
	newMessages     []*Message
	changedMessages []*Message
	voted           map[int64]bool // user hash -> voted this round
}

func NewMessages(store *SQLiteStore) *Messages {
	return &Messages{
		store:    store,
		messages: make(map[string]*Message),
		voted:    make(map[int64]bool),
	}
}

// Load populates the message structure from the database.
func (ms *Messages) Load() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.messages = make(map[string]*Message)
	ms.random = nil
	ms.newMessages = nil
	ms.changedMessages = nil

	results, err := ms.store.RetrieveMessages()
	if err != nil {
		return err
	}

	if len(results) == 0 {
		results = []*Message{{Text: "ROCK ON WITH SUPERLOUD", Author: "SUPERLOUD", Score: 1}}
	}

	for _, m := range results {
		ms.messages[m.Text] = m
	}

	// Build shuffled random queue
	keys := make([]string, 0, len(ms.messages))
	for k := range ms.messages {
		keys = append(keys, k)
	}
	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
	ms.random = keys

	ms.dirty = false
	return nil
}

// Random pops a random message, reshuffling if the queue is empty.
func (ms *Messages) Random() *Message {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if len(ms.random) == 0 {
		keys := make([]string, 0, len(ms.messages))
		for k := range ms.messages {
			keys = append(keys, k)
		}
		rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
		ms.random = keys
	}

	if len(ms.random) == 0 {
		return nil
	}

	text := ms.random[len(ms.random)-1]
	ms.random = ms.random[:len(ms.random)-1]
	ms.Last = ms.messages[text]
	ms.voted = make(map[int64]bool)

	return ms.Last
}

// Add stores the given text if it doesn't already exist.
func (ms *Messages) Add(text, author string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.messages[text]; exists {
		return
	}

	// Also check the database for messages not loaded into memory
	dbExists, err := ms.store.Exists(text)
	if err != nil {
		log.Printf("ERROR checking message existence: %v", err)
		return
	}
	if dbExists {
		return
	}

	m := &Message{
		Text:   text,
		Author: author,
		Score:  1,
		Views:  0,
	}
	ms.messages[m.Text] = m
	ms.dirty = true
	ms.newMessages = append(ms.newMessages, m)
}

// Vote casts a vote for the last message. Returns false if the user already voted.
func (ms *Messages) Vote(userHash int64, value int) bool {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.voted[userHash] {
		return false
	}

	if ms.Last == nil {
		return false
	}

	switch value {
	case 1:
		ms.Last.Score++
	case -1:
		ms.Last.Score--
	}

	ms.dirty = true
	ms.markChanged(ms.Last)
	ms.voted[userHash] = true
	return true
}

// ViewLast increments the view count for the last message.
func (ms *Messages) ViewLast() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.Last != nil {
		ms.Last.Views++
		ms.dirty = true
		ms.markChanged(ms.Last)
	}
}

// Serialize writes dirty data to the database.
func (ms *Messages) Serialize() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.dirty {
		return
	}

	for _, m := range ms.newMessages {
		id, err := ms.store.InsertMessage(m)
		if err != nil {
			log.Printf("ERROR inserting message: %v", err)
			continue
		}
		m.UID = id
	}

	for _, m := range ms.changedMessages {
		if err := ms.store.UpdateMessage(m); err != nil {
			log.Printf("ERROR updating message: %v", err)
		}
	}

	ms.dirty = false
	ms.newMessages = nil
	ms.changedMessages = nil
}

func (ms *Messages) markChanged(m *Message) {
	for _, existing := range ms.changedMessages {
		if existing == m {
			return
		}
	}
	ms.changedMessages = append(ms.changedMessages, m)
}
