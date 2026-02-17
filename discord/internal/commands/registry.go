package commands

import (
	"strings"

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

func NewRegistry(cfg *config.Config, msgs *data.Messages, filters *middleware.Filters, rpsEngine *rps.Engine) *Registry {
	r := &Registry{
		Config:   cfg,
		Messages: msgs,
		Filters:  filters,
		RPS:      rpsEngine,
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
}

// ValidCommands returns a list of all command names.
func (r *Registry) ValidCommands() []string {
	cmds := make([]string, 0, len(r.handlers))
	for k := range r.handlers {
		cmds = append(cmds, k)
	}
	return cmds
}
