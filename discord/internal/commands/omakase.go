package commands

import "strings"

func (r *Registry) cmdOmakase(ctx *CommandContext) {
	tool := "SUPERLOUD"
	if len(ctx.Params) > 0 {
		tool = strings.ToUpper(strings.Join(ctx.Params, " "))
	}

	if tool == "RAILS" || tool == "RUBY ON RAILS" {
		ctx.Reply(tool + " IS OMAKASE TIMES INFINITY AND NOT AT ALL A PROJECT THAT'S SLOWLY TURNED INTO A NIGHTMARE OF SHITTY OPINIONATED NON-ARCHITECTURE")
	} else {
		ctx.Reply(tool + " IS OMAKASE")
	}
}
