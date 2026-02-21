package commands

import (
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dialmaster/superloud-discord/internal/config"
	"github.com/dialmaster/superloud-discord/internal/data"
	"github.com/dialmaster/superloud-discord/internal/middleware"
	"github.com/dialmaster/superloud-discord/internal/rps"
)

// CommandContext abstracts over prefix messages and slash interactions.
type CommandContext struct {
	Session   *discordgo.Session
	ChannelID string
	UserID    string
	UserName  string
	Params    []string
	// Reply sends a message to the channel.
	Reply func(content string)
	// ReplyEphemeral sends an ephemeral reply (slash only; for prefix, falls back to channel).
	ReplyEphemeral func(content string)
}

type HandlerFunc func(ctx *CommandContext)

// Registry holds all command handlers and shared state.
type Registry struct {
	Config   *config.Config
	Messages *data.Messages
	Filters  *middleware.Filters
	RPS      *rps.Engine
	Store    *data.SQLiteStore

	handlers map[string]HandlerFunc

	// Dong state (daily, reset each day)
	SizeData map[int64]*SizeEntry
	Redongs  map[int64]int
}

type SizeEntry struct {
	Size int
	Nick string
	Hash int64
}

func NewRegistry(cfg *config.Config, msgs *data.Messages, filters *middleware.Filters, rpsEngine *rps.Engine, store *data.SQLiteStore) *Registry {
	r := &Registry{
		Config:   cfg,
		Messages: msgs,
		Filters:  filters,
		RPS:      rpsEngine,
		Store:    store,
		SizeData: make(map[int64]*SizeEntry),
		Redongs:  make(map[int64]int),
		handlers: make(map[string]HandlerFunc),
	}

	// Register all commands
	r.handlers["dongme"] = r.cmdDongMe
	r.handlers["redongme"] = r.cmdReDongMe
	r.handlers["sizeme"] = r.cmdSizeMe
	r.handlers["size"] = r.cmdSize
	r.handlers["biggestdong"] = r.cmdBiggestDong
	r.handlers["dongwinners"] = r.cmdDongWinners
	r.handlers["dongrankme"] = r.cmdDongRankMe
	r.handlers["dwall"] = r.cmdDWall
	r.handlers["upvote"] = r.cmdUpvote
	r.handlers["downvote"] = r.cmdDownvote
	r.handlers["score"] = r.cmdScore
	r.handlers["help"] = r.cmdHelp
	r.handlers["omakase"] = r.cmdOmakase
	r.handlers["rps"] = r.cmdRPS
	r.handlers["refresh_ignores"] = r.cmdRefreshIgnores
	r.handlers["refresh_aliases"] = r.cmdRefreshAliases
	r.handlers["undong"] = r.cmdUndong

	r.loadPersistedDongs()

	return r
}

// Dispatch handles a prefix command.
func (r *Registry) Dispatch(ctx *CommandContext, command string) {
	command = strings.ToLower(command)
	if handler, ok := r.handlers[command]; ok {
		handler(ctx)
	}
}

// IsValidCommand returns true if the command is registered.
func (r *Registry) IsValidCommand(command string) bool {
	_, ok := r.handlers[strings.ToLower(command)]
	return ok
}

// ResetDaily clears dong data for a new day.
func (r *Registry) ResetDaily() {
	r.SizeData = make(map[int64]*SizeEntry)
	r.Redongs = make(map[int64]int)
	if r.Store != nil {
		now := time.Now()
		todayInt := now.Year()*10000 + int(now.Month())*100 + now.Day()
		if err := r.Store.ClearOldDongs(todayInt); err != nil {
			log.Printf("WARNING: FAILED TO CLEAR OLD DONGS: %v", err)
		}
	}
}

// loadPersistedDongs loads today's dongs from the database into memory.
func (r *Registry) loadPersistedDongs() {
	if r.Store == nil {
		return
	}
	now := time.Now()
	todayInt := now.Year()*10000 + int(now.Month())*100 + now.Day()

	// Clean up old days
	if err := r.Store.ClearOldDongs(todayInt); err != nil {
		log.Printf("WARNING: FAILED TO CLEAR OLD DONGS: %v", err)
	}

	// Load today's dongs
	entries, redongs, err := r.Store.LoadDongs(todayInt)
	if err != nil {
		log.Printf("WARNING: FAILED TO LOAD PERSISTED DONGS: %v", err)
		return
	}

	for hash, entry := range entries {
		r.SizeData[hash] = &SizeEntry{
			Size: entry.Size,
			Nick: entry.Nick,
			Hash: entry.Hash,
		}
	}
	for hash, count := range redongs {
		r.Redongs[hash] = count
	}

	if len(entries) > 0 {
		log.Printf("LOADED %d PERSISTED DONGS FOR TODAY", len(entries))
	}
}

// ValidCommands returns a list of all command names.
func (r *Registry) ValidCommands() []string {
	cmds := make([]string, 0, len(r.handlers))
	for k := range r.handlers {
		cmds = append(cmds, k)
	}
	return cmds
}
