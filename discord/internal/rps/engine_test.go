package rps

import "testing"

func newTestEngine() *Engine {
	return &Engine{
		Name:    "Test RPS",
		Objects: []string{"rock", "paper", "scissors"},
		Messages: map[string]map[string]string{
			"rock":     {"scissors": "Rock smashes scissors"},
			"paper":    {"rock": "Paper covers rock"},
			"scissors": {"paper": "Scissors cuts paper"},
		},
	}
}

func TestFight_AttackerWins(t *testing.T) {
	e := newTestEngine()
	wins, tie, msg := e.Fight("rock", "scissors")
	if !wins || tie {
		t.Errorf("expected rock to beat scissors")
	}
	if msg != "Rock smashes scissors" {
		t.Errorf("unexpected message: %s", msg)
	}
}

func TestFight_DefenderWins(t *testing.T) {
	e := newTestEngine()
	wins, tie, msg := e.Fight("scissors", "rock")
	if wins || tie {
		t.Errorf("expected scissors to lose to rock")
	}
	if msg != "Rock smashes scissors" {
		t.Errorf("unexpected message: %s", msg)
	}
}

func TestFight_Tie(t *testing.T) {
	e := newTestEngine()
	_, tie, _ := e.Fight("rock", "rock")
	if !tie {
		t.Errorf("expected tie when same objects fight")
	}
}

func TestValidObject(t *testing.T) {
	e := newTestEngine()
	if !e.ValidObject("rock") {
		t.Error("rock should be valid")
	}
	if !e.ValidObject("ROCK") {
		t.Error("ROCK should be valid (case insensitive)")
	}
	if e.ValidObject("banana") {
		t.Error("banana should not be valid")
	}
}

func TestObjectList(t *testing.T) {
	e := newTestEngine()
	list := e.ObjectList()
	if len(list) != 3 {
		t.Errorf("expected 3 objects, got %d", len(list))
	}
}
