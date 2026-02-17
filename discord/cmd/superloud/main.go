package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
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
)

var (
	cfg      *config.Config
	msgs     *data.Messages
	registry *commands.Registry
	filters  *middleware.Filters
	lastDay  int
)

func main() {
	// Load config
	configPath := "config/config.yml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	var err error
	cfg, err = config.Load(configPath)
	if err != nil {
		log.Fatalf("FAILED TO LOAD CONFIG: %v", err)
	}

	if cfg.Token == "" {
		log.Fatal("DISCORD_TOKEN ENVIRONMENT VARIABLE IS NOT SET")
	}

	// Initialize SQLite store
	store, err := data.NewSQLiteStore(cfg.DBPath)
	if err != nil {
		log.Fatalf("FAILED TO OPEN DATABASE: %v", err)
	}
	defer store.Close()

	// Initialize messages
	msgs = data.NewMessages(store)
	if err := msgs.Load(); err != nil {
		log.Fatalf("FAILED TO LOAD MESSAGES: %v", err)
	}

	// Initialize filters
	filters = middleware.NewFilters()
	if err := filters.LoadIgnores(cfg); err != nil {
		log.Printf("WARNING: FAILED TO LOAD IGNORES: %v", err)
	}
	if err := filters.LoadAliases(cfg); err != nil {
		log.Printf("WARNING: FAILED TO LOAD ALIASES: %v", err)
	}

	// Initialize RPS engine
	rpsEngine, err := rps.LoadEngine(cfg.RPSPath)
	if err != nil {
		log.Printf("WARNING: FAILED TO LOAD RPS CONFIG: %v", err)
	}

	// Initialize command registry
	registry = commands.NewRegistry(cfg, msgs, filters, rpsEngine)

	// Initialize daily data
	lastDay = todayInt()

	// Create Discord session
	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		log.Fatalf("FAILED TO CREATE DISCORD SESSION: %v", err)
	}

	// Set intents
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent | discordgo.IntentsGuilds

	// Register event handlers
	dg.AddHandler(onReady)
	dg.AddHandler(onMessageCreate)
	dg.AddHandler(onInteractionCreate)

	// Open connection
	if err := dg.Open(); err != nil {
		log.Fatalf("FAILED TO OPEN DISCORD CONNECTION: %v", err)
	}
	defer dg.Close()

	// Register slash commands
	registerSlashCommands(dg)

	// Start periodic serialization ticker
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			msgs.Serialize()
			today := todayInt()
			if lastDay != today {
				registry.ResetDaily()
				lastDay = today
			}
		}
	}()

	// Wait for interrupt signal
	fmt.Println("SUPERLOUD IS NOW RUNNING. PRESS CTRL+C TO EXIT.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	fmt.Println("SHUTTING DOWN...")
	msgs.Serialize()
}

func todayInt() int {
	now := time.Now()
	return now.Year()*10000 + int(now.Month())*100 + now.Day()
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("LOGGED IN AS %s", r.User.Username)
	if cfg.ChannelID != "" {
		s.ChannelMessageSend(cfg.ChannelID, "WHATS WRONG WITH BEING SEXY")
	}
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
		return
	}

	// Check whitelist
	if !filters.IsWhitelisted(userID) {
		return
	}

	// Resolve aliases
	resolvedID := filters.ResolveAlias(userID)
	_ = resolvedID

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
			sendRandomMessage(s, m.ChannelID)
			return
		}
	}

	// Check if the message is loud
	status, _ := loudbot.IsItLoud(text)
	switch status {
	case loudbot.StatusBad:
		return
	case loudbot.StatusRejected:
		sendRandomMessage(s, m.ChannelID)
	case loudbot.StatusLoud:
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
		command = cmdData.Options[0].Name
		params = nil // admin subcommands don't have additional params
	}

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
			},
		},
	}

	for _, cmd := range cmds {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("WARNING: FAILED TO REGISTER SLASH COMMAND %s: %v", cmd.Name, err)
		}
	}
}
