package middleware

import (
	"os"
	"regexp"
	"sync"

	"github.com/dialmaster/superloud-discord/internal/config"
	"gopkg.in/yaml.v3"
)

type Filters struct {
	mu               sync.RWMutex
	ignoreRegexes    []*regexp.Regexp
	whitelistRegexes []*regexp.Regexp
	aliases          map[string]string // secondary user ID -> primary user ID
}

func NewFilters() *Filters {
	return &Filters{
		aliases: make(map[string]string),
	}
}

// LoadIgnores loads the ignore and whitelist patterns from config files.
func (f *Filters) LoadIgnores(cfg *config.Config) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.ignoreRegexes = nil
	f.whitelistRegexes = nil

	ignores, err := config.LoadLines(cfg.IgnoresPath)
	if err != nil {
		return err
	}
	for _, pattern := range ignores {
		re, err := regexp.Compile("(?i)" + pattern)
		if err != nil {
			continue
		}
		f.ignoreRegexes = append(f.ignoreRegexes, re)
	}

	allowed, err := config.LoadLines(cfg.WhitelistPath)
	if err != nil {
		return err
	}
	for _, pattern := range allowed {
		re, err := regexp.Compile("(?i)" + pattern)
		if err != nil {
			continue
		}
		f.whitelistRegexes = append(f.whitelistRegexes, re)
	}

	return nil
}

type aliasFile struct {
	Aliases map[string]string `yaml:"aliases"`
}

// LoadAliases loads the alias mappings from the YAML config file.
func (f *Filters) LoadAliases(cfg *config.Config) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.aliases = make(map[string]string)

	data, err := os.ReadFile(cfg.AliasesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var af aliasFile
	if err := yaml.Unmarshal(data, &af); err != nil {
		return err
	}

	for secondary, primary := range af.Aliases {
		f.aliases[secondary] = primary
	}

	return nil
}

// ResolveAlias returns the primary user ID for the given user ID.
// If no alias exists, returns the original user ID.
func (f *Filters) ResolveAlias(userID string) string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if primary, ok := f.aliases[userID]; ok {
		return primary
	}
	return userID
}

// ShouldIgnore checks if a user ID should be ignored.
func (f *Filters) ShouldIgnore(userID string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, re := range f.ignoreRegexes {
		if re.MatchString(userID) {
			return true
		}
	}
	return false
}

// IsWhitelisted checks if a user passes the whitelist.
// Returns true if no whitelist is configured or the user matches.
func (f *Filters) IsWhitelisted(userID string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if len(f.whitelistRegexes) == 0 {
		return true
	}

	for _, re := range f.whitelistRegexes {
		if re.MatchString(userID) {
			return true
		}
	}
	return false
}
