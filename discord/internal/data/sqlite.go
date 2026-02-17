package data

import (
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"os"

	_ "modernc.org/sqlite"
)

const (
	minimumRetrievalSize = 100
	maximumRetrievalSize = minimumRetrievalSize * 100
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(filename string) (*SQLiteStore, error) {
	dbExists := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		dbExists = false
	}

	db, err := sql.Open("sqlite", filename)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if !dbExists {
		schema := `
			CREATE TABLE messages (
				id INTEGER PRIMARY KEY,
				text TEXT,
				author TEXT,
				score INTEGER,
				views INTEGER
			);
			CREATE UNIQUE INDEX messages_text_idx ON messages (text);
			CREATE INDEX messages_author_idx ON messages (author);
			CREATE INDEX messages_score_idx ON messages (score);
			CREATE INDEX messages_views_idx ON messages (views);
		`
		if _, err := db.Exec(schema); err != nil {
			db.Close()
			return nil, fmt.Errorf("creating schema: %w", err)
		}
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// Exists checks if a message with the given text exists in the database.
func (s *SQLiteStore) Exists(text string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM messages WHERE text = ?", text).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// InsertMessage adds a new message to the database and returns its ID.
func (s *SQLiteStore) InsertMessage(m *Message) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO messages (text, score, author, views) VALUES (?, ?, ?, ?)",
		m.Text, m.Score, m.Author, m.Views,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateMessage updates score and views for a message by ID.
func (s *SQLiteStore) UpdateMessage(m *Message) error {
	_, err := s.db.Exec(
		"UPDATE messages SET score = ?, views = ? WHERE id = ?",
		m.Score, m.Views, m.UID,
	)
	return err
}

// RetrieveMessages implements the weighted random retrieval algorithm from the Ruby version.
func (s *SQLiteStore) RetrieveMessages() ([]*Message, error) {
	rows, err := s.db.Query(`
		SELECT id, text, score, author, views
		FROM messages WHERE score > -1
		ORDER BY views
		LIMIT ?
	`, maximumRetrievalSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allResults []*Message
	for rows.Next() {
		m := &Message{}
		if err := rows.Scan(&m.UID, &m.Text, &m.Score, &m.Author, &m.Views); err != nil {
			return nil, err
		}
		allResults = append(allResults, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Shuffle since we only used ORDER BY views to ensure new items get in
	rand.Shuffle(len(allResults), func(i, j int) {
		allResults[i], allResults[j] = allResults[j], allResults[i]
	})

	// Return full set if we didn't pull many beyond the bare minimum
	if len(allResults) <= minimumRetrievalSize*2 {
		return allResults, nil
	}

	// Weighted random selection
	baseChance := 3.0 * (float64(minimumRetrievalSize) / float64(len(allResults)))

	var results []*Message
	for _, m := range allResults {
		weight := baseChance * (math.Pow(1.0+float64(m.Score), 0.75) / float64(m.Views+1))

		if m.Views == 0 {
			weight *= 2
		}

		if rand.Float64() < weight {
			results = append(results, m)
		}
	}

	return results, nil
}
