package commands

import (
	"fmt"
	"strings"
	"sync"
)

// RPSContestant holds a registered RPS player for a channel.
type RPSContestant struct {
	UserID   string
	UserName string
	Object   string
}

var (
	rpsContestants   = make(map[string]*RPSContestant) // channelID -> contestant
	rpsContestantsMu sync.Mutex
)

func (r *Registry) cmdRPS(ctx *CommandContext) {
	if r.RPS == nil {
		ctx.Reply("RPS IS NOT CONFIGURED")
		return
	}

	if len(ctx.Params) == 0 {
		ctx.Reply(fmt.Sprintf("OBJECT MUST BE ONE OF THE FOLLOWING: %s", strings.Join(r.RPS.ObjectList(), ", ")))
		return
	}

	object := strings.ToLower(ctx.Params[0])

	if !r.RPS.ValidObject(object) {
		ctx.Reply(fmt.Sprintf("HEY DUMBFACE THAT IS AN INVALID OBJECT. OBJECT MUST BE ONE OF THE FOLLOWING: %s", strings.Join(r.RPS.ObjectList(), ", ")))
		return
	}

	// Must be loud
	if strings.ToUpper(ctx.Params[0]) != ctx.Params[0] {
		ctx.Reply(fmt.Sprintf("IT SOUNDED LIKE YOU REQUESTED A %s BUT I CAN'T QUITE HEAR YOU", strings.ToUpper(ctx.Params[0])))
		return
	}

	rpsContestantsMu.Lock()
	defer rpsContestantsMu.Unlock()

	rpsName := strings.ToUpper(r.RPS.Name)

	// Self-fight detection (check before confirming)
	existing := rpsContestants[ctx.ChannelID]
	if existing != nil && existing.UserID == ctx.UserID {
		ctx.ReplyEphemeral("GET THE FUCK OUT OF MY FACE THIS IS NOT FIGHT CLUB YOU CANNOT FIGHT YOURSELF")
		return
	}

	challenger := &RPSContestant{
		UserID:   ctx.UserID,
		UserName: ctx.UserName,
		Object:   object,
	}

	// No existing contestant - register as first player
	if existing == nil {
		rpsContestants[ctx.ChannelID] = challenger
		ctx.ReplyEphemeral("OKAY I WILL SET YOU UP THE BOMB ALSO THX 4 PLAYING")
		ctx.Session.ChannelMessageSend(ctx.ChannelID, fmt.Sprintf("%s HAS REGISTERED FOR %s!  WHO IS BRAVE ENOUGH TO FIGHT???", strings.ToUpper(ctx.UserName), rpsName))
		return
	}

	// Second player - run the fight
	ctx.ReplyEphemeral("OKAY I WILL SET YOU UP THE BOMB ALSO THX 4 PLAYING")
	ctx.Session.ChannelMessageSend(ctx.ChannelID, fmt.Sprintf("%s HAS REGISTERED FOR %s TO CHALLENGE %s...", strings.ToUpper(ctx.UserName), rpsName, strings.ToUpper(existing.UserName)))

	attackerWins, tie, fightMsg := r.RPS.Fight(challenger.Object, existing.Object)

	var text string
	switch {
	case tie:
		text = fmt.Sprintf("BUT BOTH WERE USING %s.  OMFG HOW GAY A TIE.", strings.ToUpper(challenger.Object))
	case attackerWins:
		text = fmt.Sprintf("AND DEFEATS %s: %s", strings.ToUpper(existing.UserName), fightMsg)
	default:
		text = fmt.Sprintf("BUT %s IS DEFEATED: %s", strings.ToUpper(challenger.UserName), fightMsg)
	}

	ctx.Session.ChannelMessageSend(ctx.ChannelID, text)

	// Clear the contestant for this channel
	delete(rpsContestants, ctx.ChannelID)
}
