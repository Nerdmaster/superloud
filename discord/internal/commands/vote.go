package commands

import (
	"fmt"

	"github.com/dialmaster/superloud-discord/internal/util"
)

func (r *Registry) cmdUpvote(ctx *CommandContext) {
	userHash := util.UserHash(r.Filters.ResolveAlias(ctx.UserID))
	if !r.Messages.Vote(userHash, 1) {
		ctx.Reply("SORRY YOU CAN'T VOTE ON THIS MESSAGE AGAIN")
	}
}

func (r *Registry) cmdDownvote(ctx *CommandContext) {
	userHash := util.UserHash(r.Filters.ResolveAlias(ctx.UserID))
	if !r.Messages.Vote(userHash, -1) {
		ctx.Reply("SORRY YOU CAN'T VOTE ON THIS MESSAGE AGAIN")
	}
}

func (r *Registry) cmdScore(ctx *CommandContext) {
	last := r.Messages.Last
	if last == nil {
		ctx.Reply("NO LAST MESSAGE OR IT WAS DELETED BY !DOWNVOTE")
		return
	}

	ctx.Reply(fmt.Sprintf("%s: %d, SUBMITTED BY %s", last.Text, last.Score, last.Author))
}
