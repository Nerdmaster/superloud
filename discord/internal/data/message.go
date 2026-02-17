package data

// Message represents a single loud message from the database.
type Message struct {
	UID    int64
	Text   string
	Author string
	Score  int
	Views  int
}
