package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dialmaster/superloud-discord/internal/commands"
	"github.com/dialmaster/superloud-discord/internal/config"
	"github.com/dialmaster/superloud-discord/internal/data"
	"github.com/dialmaster/superloud-discord/internal/loudbot"
	"github.com/dialmaster/superloud-discord/internal/middleware"
	"github.com/dialmaster/superloud-discord/internal/rps"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	cfg      *config.Config
	msgs     *data.Messages
	registry *commands.Registry
	filters  *middleware.Filters
	lastDay  int
)

func initLogger(cfg *config.Config) {
	// Parse log level
	var level slog.Level
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Ensure log directory exists
	logDir := filepath.Dir(cfg.LogPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "FAILED TO CREATE LOG DIRECTORY %s: %v\n", logDir, err)
		os.Exit(1)
	}

	// Create lumberjack rotating writer
	fileWriter := &lumberjack.Logger{
		Filename:   cfg.LogPath,
		MaxSize:    cfg.LogMaxSizeMB,
		MaxBackups: cfg.LogMaxBackups,
		MaxAge:     cfg.LogMaxAgeDays,
		Compress:   cfg.LogCompress,
	}

	// Write to both file and stdout
	multiWriter := io.MultiWriter(fileWriter, os.Stdout)

	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))
}

func main() {
	// Load config
	configPath := "config/config.yml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	var err error
	cfg, err = config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAILED TO LOAD CONFIG: %v\n", err)
		os.Exit(1)
	}

	if cfg.Token == "" {
		fmt.Fprintf(os.Stderr, "DISCORD_TOKEN ENVIRONMENT VARIABLE IS NOT SET\n")
		os.Exit(1)
	}

	// Initialize logger (requires config)
	initLogger(cfg)
	slog.Info("configuration loaded", "config_path", configPath)

	// Initialize SQLite store
	store, err := data.NewSQLiteStore(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open database", "db_path", cfg.DBPath, "error", err)
		os.Exit(1)
	}
	defer store.Close()
	slog.Info("database opened", "db_path", cfg.DBPath)

	// Initialize messages
	msgs = data.NewMessages(store)
	if err := msgs.Load(); err != nil {
		slog.Error("failed to load messages", "error", err)
		os.Exit(1)
	}
	slog.Info("messages loaded")

	// Initialize filters
	filters = middleware.NewFilters()
	if err := filters.LoadIgnores(cfg); err != nil {
		slog.Warn("failed to load ignores", "error", err)
	}
	if err := filters.LoadAliases(cfg); err != nil {
		slog.Warn("failed to load aliases", "error", err)
	}
	slog.Info("filters initialized")

	// Initialize RPS engine
	rpsEngine, err := rps.LoadEngine(cfg.RPSPath)
	if err != nil {
		slog.Warn("failed to load RPS config", "rps_path", cfg.RPSPath, "error", err)
	} else {
		slog.Info("RPS engine loaded", "rps_path", cfg.RPSPath)
	}

	// Initialize command registry
	registry = commands.NewRegistry(cfg, msgs, filters, rpsEngine, store)
	slog.Info("command registry initialized")

	// Initialize daily data
	lastDay = todayInt()

	// Create Discord session
	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		slog.Error("failed to create Discord session", "error", err)
		os.Exit(1)
	}

	// Set intents
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent | discordgo.IntentsGuilds

	// Register event handlers
	dg.AddHandler(onReady)
	dg.AddHandler(onMessageCreate)
	dg.AddHandler(onInteractionCreate)

	// Open connection
	if err := dg.Open(); err != nil {
		slog.Error("failed to open Discord connection", "error", err)
		os.Exit(1)
	}
	defer dg.Close()
	slog.Info("Discord connection opened")

	// Register slash commands
	registerSlashCommands(dg)

	// Start periodic serialization ticker
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			slog.Debug("periodic serialization tick")
			msgs.Serialize()
			today := todayInt()
			if lastDay != today {
				slog.Info("daily reset triggered", "old_day", lastDay, "new_day", today)
				registry.ResetDaily()
				lastDay = today
			}
		}
	}()

	// Wait for interrupt signal
	slog.Info("superloud is now running, waiting for shutdown signal")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sc

	slog.Info("shutdown signal received", "signal", sig.String())
	msgs.Serialize()
	slog.Info("shutdown complete")
}

func todayInt() int {
	now := time.Now()
	return now.Year()*10000 + int(now.Month())*100 + now.Day()
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	slog.Info("logged in", "username", r.User.Username, "user_id", r.User.ID)
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore DMs
	if m.GuildID == "" {
		return
	}

	userID := m.Author.ID
	userName := m.Author.Username
	if m.Member != nil && m.Member.Nick != "" {
		userName = m.Member.Nick
	}

	// Check ignore list
	if filters.ShouldIgnore(userID) {
		slog.Debug("message from ignored user", "user_id", userID, "username", userName)
		return
	}

	// Check whitelist
	if !filters.IsWhitelisted(userID) {
		slog.Debug("message from non-whitelisted user", "user_id", userID, "username", userName)
		return
	}

	// Resolve aliases
	resolvedID := filters.ResolveAlias(userID)
	if resolvedID != userID {
		slog.Debug("alias resolved", "user_id", userID, "resolved_id", resolvedID)
	}

	text := m.Content

	// Check for prefix commands (!)
	if strings.HasPrefix(text, "!") {
		parts := strings.Fields(text)
		command := strings.TrimPrefix(parts[0], "!")

		// Commands must be uppercase
		if command != strings.ToUpper(strings.TrimSpace(command)) {
			return
		}

		if registry.IsValidCommand(command) {
			slog.Info("prefix command received",
				"command", command,
				"user_id", userID,
				"username", userName,
				"channel_id", m.ChannelID,
				"params", parts[1:],
			)
			ctx := &commands.CommandContext{
				Session:   s,
				ChannelID: m.ChannelID,
				UserID:    userID,
				UserName:  userName,
				Params:    parts[1:],
				Reply: func(content string) {
					s.ChannelMessageSend(m.ChannelID, content)
				},
				ReplyEphemeral: func(content string) {
					s.ChannelMessageSend(m.ChannelID, content)
				},
			}
			registry.Dispatch(ctx, command)
			return
		}
	}

	// Check if someone is talking to the bot (mentions)
	for _, mention := range m.Mentions {
		if mention.ID == s.State.User.ID {
			slog.Debug("bot mentioned", "user_id", userID, "username", userName, "channel_id", m.ChannelID)
			sendRandomMessage(s, m.ChannelID)
			return
		}
	}

	// Check if the message is loud
	status, reason := loudbot.IsItLoud(text)
	switch status {
	case loudbot.StatusBad:
		return
	case loudbot.StatusRejected:
		slog.Debug("loud message rejected", "reason", reason, "user_id", userID, "username", userName)
		sendRandomMessage(s, m.ChannelID)
	case loudbot.StatusLoud:
		slog.Info("loud message accepted", "user_id", userID, "username", userName, "channel_id", m.ChannelID)
		sendRandomMessage(s, m.ChannelID)
		msgs.Add(text, userName)
	}
}

func sendRandomMessage(s *discordgo.Session, channelID string) {
	msg := msgs.Random()
	if msg == nil {
		return
	}
	s.ChannelMessageSend(channelID, msg.Text)
	msgs.ViewLast()
}

func onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	cmdData := i.ApplicationCommandData()
	userName := ""
	userID := ""

	if i.Member != nil {
		userID = i.Member.User.ID
		userName = i.Member.User.Username
		if i.Member.Nick != "" {
			userName = i.Member.Nick
		}
	} else if i.User != nil {
		userID = i.User.ID
		userName = i.User.Username
	}

	// Parse options into params
	var params []string
	for _, opt := range cmdData.Options {
		switch opt.Type {
		case discordgo.ApplicationCommandOptionString:
			params = append(params, opt.StringValue())
		case discordgo.ApplicationCommandOptionInteger:
			params = append(params, fmt.Sprintf("%d", opt.IntValue()))
		case discordgo.ApplicationCommandOptionUser:
			// For user mentions, get the user's display name
			if opt.UserValue(s) != nil {
				params = append(params, opt.UserValue(s).Username)
			}
		}
	}

	// Handle admin subcommand
	command := cmdData.Name
	if command == "admin" && len(cmdData.Options) > 0 {
		subCmd := cmdData.Options[0]
		command = subCmd.Name
		// Parse subcommand options
		params = nil
		for _, opt := range subCmd.Options {
			switch opt.Type {
			case discordgo.ApplicationCommandOptionUser:
				if opt.UserValue(s) != nil {
					params = append(params, opt.UserValue(s).Username)
				}
			case discordgo.ApplicationCommandOptionString:
				params = append(params, opt.StringValue())
			case discordgo.ApplicationCommandOptionInteger:
				params = append(params, fmt.Sprintf("%d", opt.IntValue()))
			}
		}
	}

	slog.Info("slash command received",
		"command", command,
		"user_id", userID,
		"username", userName,
		"channel_id", i.ChannelID,
		"params", params,
	)

	// Determine if this should be ephemeral
	isEphemeral := command == "rps"

	ctx := &commands.CommandContext{
		Session:   s,
		ChannelID: i.ChannelID,
		UserID:    userID,
		UserName:  userName,
		Params:    params,
		Reply: func(content string) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			})
		},
		ReplyEphemeral: func(content string) {
			flags := discordgo.MessageFlags(0)
			if isEphemeral {
				flags = discordgo.MessageFlagsEphemeral
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
					Flags:   flags,
				},
			})
		},
	}

	registry.Dispatch(ctx, command)
}

func registerSlashCommands(s *discordgo.Session) {
	cmds := []*discordgo.ApplicationCommand{
		{
			Name:        "dongme",
			Description: "SHOWS HOW MUCH OF A MAN YOU ARE",
		},
		{
			Name:        "redongme",
			Description: "REROLL YOUR DONG WITH DANGER RISK",
		},
		{
			Name:        "sizeme",
			Description: "TELLS YOU IF YOU ARE WORTH ANYTHING TO SOCIETY",
		},
		{
			Name:        "size",
			Description: "GIVES YOU THE ONLY THING THAT MATTERS ABOUT SOMEBODY: SIZE",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "THE USER TO CHECK",
					Required:    true,
				},
			},
		},
		{
			Name:        "biggestdong",
			Description: "WHO HAS THE BIGGEST DONG TODAY",
		},
		{
			Name:        "dongwinners",
			Description: "SHOW THE PEOPLE WHO FUCKING MATTER",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "count",
					Description: "NUMBER OF PLACES TO SHOW (DEFAULT 2)",
					Required:    false,
				},
			},
		},
		{
			Name:        "dongrankme",
			Description: "SHOW YOUR RELATIVE WORTH",
		},
		{
			Name:        "dwall",
			Description: "SHOW EVERYBODY'S RANK",
		},
		{
			Name:        "upvote",
			Description: "VOTE THE CURRENT MESSAGE +1",
		},
		{
			Name:        "downvote",
			Description: "VOTE THE CURRENT MESSAGE -1",
		},
		{
			Name:        "score",
			Description: "SHOW THE LAST MESSAGE'S SCORE",
		},
		{
			Name:        "help",
			Description: "GET HELP WITH COMMANDS",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "command",
					Description: "SPECIFIC COMMAND TO GET HELP FOR",
					Required:    false,
				},
			},
		},
		{
			Name:        "omakase",
			Description: "MAKES TOOLS REALLY GREAT",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "tool",
					Description: "THE TOOL TO JUDGE",
					Required:    false,
				},
			},
		},
		{
			Name:        "rps",
			Description: "PLAY ROCK PAPER SCISSORS (EXTENDED EDITION)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "object",
					Description: "YOUR WEAPON OF CHOICE",
					Required:    true,
				},
			},
		},
		{
			Name:        "admin",
			Description: "ADMIN COMMANDS",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "refresh_ignores",
					Description: "RELOAD THE IGNORE LIST",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "refresh_aliases",
					Description: "RELOAD THE ALIAS LIST",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "undong",
					Description: "REMOVE A USER'S DONG FOR TODAY",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "THE USER WHOSE DONG TO REMOVE",
							Required:    true,
						},
					},
				},
			},
		},
	}

	for _, cmd := range cmds {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			slog.Error("failed to register slash command", "command", cmd.Name, "error", err)
		} else {
			slog.Debug("slash command registered", "command", cmd.Name)
		}
	}
}
