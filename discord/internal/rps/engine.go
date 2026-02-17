package rps

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Engine struct {
	Name     string
	Objects  []string
	Messages map[string]map[string]string // attacker -> defender -> message
}

type rpsConfig struct {
	Name     string                       `yaml:":name"`
	Messages map[string]map[string]string `yaml:":messages"`
}

// LoadEngine loads the RPS configuration from a YAML file.
func LoadEngine(filename string) (*Engine, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading RPS config: %w", err)
	}

	var raw rpsConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing RPS config: %w", err)
	}

	e := &Engine{
		Name:     raw.Name,
		Messages: make(map[string]map[string]string),
	}

	for objKey, battles := range raw.Messages {
		// Strip leading colon from YAML symbol keys
		obj := strings.TrimPrefix(objKey, ":")
		e.Objects = append(e.Objects, obj)

		e.Messages[obj] = make(map[string]string)
		for defKey, msg := range battles {
			def := strings.TrimPrefix(defKey, ":")
			e.Messages[obj][def] = msg
		}
	}

	return e, nil
}

// ValidObject checks if the given object name is valid.
func (e *Engine) ValidObject(obj string) bool {
	obj = strings.ToLower(obj)
	for _, o := range e.Objects {
		if o == obj {
			return true
		}
	}
	return false
}

// ObjectList returns all object names in uppercase.
func (e *Engine) ObjectList() []string {
	result := make([]string, len(e.Objects))
	for i, o := range e.Objects {
		result[i] = strings.ToUpper(o)
	}
	return result
}

// Fight determines the winner between two objects.
// Returns: true if attacker wins, false if defender wins, nil-equivalent (tie) indicated by third return.
func (e *Engine) Fight(attacker, defender string) (attackerWins bool, tie bool, message string) {
	attacker = strings.ToLower(attacker)
	defender = strings.ToLower(defender)

	if attacker == defender {
		return false, true, "TIE!"
	}

	// Check if attacker has a win message against defender
	if msgs, ok := e.Messages[attacker]; ok {
		if msg, ok := msgs[defender]; ok {
			return true, false, msg
		}
	}

	// Defender wins - get the message
	if msgs, ok := e.Messages[defender]; ok {
		if msg, ok := msgs[attacker]; ok {
			return false, false, msg
		}
	}

	return false, true, "NO BATTLE CONFIGURED BETWEEN THESE OBJECTS"
}
