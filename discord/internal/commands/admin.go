package commands

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dialmaster/superloud-discord/internal/util"
)

func (r *Registry) cmdUndong(ctx *CommandContext) {
	if !r.Config.IsAdmin(ctx.UserID) {
		return
	}

	if len(ctx.Params) == 0 {
		ctx.Reply("USAGE: !UNDONG <USERNAME>")
		return
	}

	target := ctx.Params[0]

	// Strip Discord mention formatting: <@123456> or <@!123456>
	cleanTarget := target
	cleanTarget = strings.TrimPrefix(cleanTarget, "<@")
	cleanTarget = strings.TrimPrefix(cleanTarget, "!")
	cleanTarget = strings.TrimSuffix(cleanTarget, ">")

	// Try lookup by user ID hash first (works for mentions)
	var foundHash int64
	var foundNick string

	if cleanTarget != target || isNumeric(cleanTarget) {
		// Looks like a user ID - try hash lookup
		hash := util.UserHash(r.Filters.ResolveAlias(cleanTarget))
		if entry, ok := r.SizeData[hash]; ok {
			foundHash = hash
			foundNick = entry.Nick
		}
	}

	// Fall back to case-insensitive nick search
	if foundHash == 0 {
		upperTarget := strings.ToUpper(cleanTarget)
		for hash, entry := range r.SizeData {
			if strings.ToUpper(entry.Nick) == upperTarget {
				foundHash = hash
				foundNick = entry.Nick
				break
			}
		}
	}

	if foundHash == 0 {
		ctx.Reply(fmt.Sprintf("NO DONG FOUND FOR %s", strings.ToUpper(cleanTarget)))
		return
	}

	// Delete from memory
	delete(r.SizeData, foundHash)
	delete(r.Redongs, foundHash)

	// Delete from DB
	if r.Store != nil {
		todayInt := dateToInt(time.Now())
		if err := r.Store.DeleteDong(todayInt, foundHash); err != nil {
			log.Printf("WARNING: FAILED TO DELETE DONG FROM DB: %v", err)
		}
	}

	ctx.Reply(fmt.Sprintf("%s'S DONG HAS BEEN REMOVED", strings.ToUpper(foundNick)))
}

// isNumeric returns true if the string contains only digits.
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func (r *Registry) cmdRefreshIgnores(ctx *CommandContext) {
	if !r.Config.IsAdmin(ctx.UserID) {
		return
	}

	if err := r.Filters.LoadIgnores(r.Config); err != nil {
		ctx.Reply("ERROR RELOADING IGNORES: " + err.Error())
		return
	}
	ctx.Reply("IGNORES RELOADED SUCCESSFULLY")
}

func (r *Registry) cmdRefreshAliases(ctx *CommandContext) {
	if !r.Config.IsAdmin(ctx.UserID) {
		return
	}

	if err := r.Filters.LoadAliases(r.Config); err != nil {
		ctx.Reply("ERROR RELOADING ALIASES: " + err.Error())
		return
	}
	ctx.Reply("ALIASES RELOADED SUCCESSFULLY")
}
