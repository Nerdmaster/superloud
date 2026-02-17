package commands

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
