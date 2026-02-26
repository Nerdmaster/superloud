package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Token         string   `yaml:"-"`
	ChannelID     string   `yaml:"channel_id"`
	AdminIDs      []string `yaml:"admin_ids"`
	DBPath        string   `yaml:"db_path"`
	RPSPath       string   `yaml:"rps_path"`
	IgnoresPath   string   `yaml:"ignores_path"`
	WhitelistPath string   `yaml:"whitelist_path"`
	AliasesPath   string   `yaml:"aliases_path"`
	LogPath       string   `yaml:"log_path"`
	LogMaxSizeMB  int      `yaml:"log_max_size_mb"`
	LogMaxBackups int      `yaml:"log_max_backups"`
	LogMaxAgeDays int      `yaml:"log_max_age_days"`
	LogCompress   bool     `yaml:"log_compress"`
	LogLevel      string   `yaml:"log_level"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		DBPath:        "db/louds.db",
		RPSPath:       "rps/rps.yml",
		IgnoresPath:   "config/ignores.txt",
		WhitelistPath: "config/whitelist.txt",
		AliasesPath:   "config/aliases.yml",
		LogPath:       "logs/superloud.log",
		LogMaxSizeMB:  50,
		LogMaxBackups: 3,
		LogMaxAgeDays: 30,
		LogCompress:   true,
		LogLevel:      "info",
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	cfg.Token = os.Getenv("DISCORD_TOKEN")

	return cfg, nil
}

func (c *Config) IsAdmin(userID string) bool {
	for _, id := range c.AdminIDs {
		if id == userID {
			return true
		}
	}
	return false
}

// LoadLines reads a file and returns non-empty trimmed lines.
func LoadLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}
