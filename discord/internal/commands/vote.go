package commands

import (
	"fmt"
	"log/slog"

	"github.com/dialmaster/superloud-discord/internal/util"
)

func (r *Registry) cmdUpvote(ctx *CommandContext) {
	userHash := util.UserHash(r.Filters.ResolveAlias(ctx.UserID))
	if !r.Messages.Vote(userHash, 1) {
		slog.Debug("duplicate vote rejected", "user_id", ctx.UserID, "username", ctx.UserName, "direction", "up")
		ctx.Reply("SORRY YOU CAN'T VOTE ON THIS MESSAGE AGAIN")
	} else {
		slog.Debug("upvote cast", "user_id", ctx.UserID, "username", ctx.UserName)
	}
}

func (r *Registry) cmdDownvote(ctx *CommandContext) {
	userHash := util.UserHash(r.Filters.ResolveAlias(ctx.UserID))
	if !r.Messages.Vote(userHash, -1) {
		slog.Debug("duplicate vote rejected", "user_id", ctx.UserID, "username", ctx.UserName, "direction", "down")
		ctx.Reply("SORRY YOU CAN'T VOTE ON THIS MESSAGE AGAIN")
	} else {
		slog.Debug("downvote cast", "user_id", ctx.UserID, "username", ctx.UserName)
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
