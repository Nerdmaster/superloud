package commands

import (
	"sort"
	"strings"
)

func (r *Registry) cmdHelp(ctx *CommandContext) {
	if len(ctx.Params) == 0 {
		cmds := r.ValidCommands()
		sort.Strings(cmds)
		cmdList := make([]string, len(cmds))
		for i, c := range cmds {
			cmdList[i] = "!" + strings.ToUpper(c)
		}
		ctx.Reply("I HAVE COMMANDS AND THEY ARE: " + strings.Join(cmdList, " "))
		return
	}

	if len(ctx.Params) > 1 {
		ctx.Reply("WTF ARE YOU DUMB?  I OFFER HELP FOR ONE COMMAND AT A TIME JERKFACE")
		return
	}

	cmd := strings.ToUpper(ctx.Params[0])
	if !r.IsValidCommand(cmd) {
		ctx.Reply("!" + cmd + " IS NOT A COMMAND YOU TWIT")
		return
	}

	switch cmd {
	case "DWALL":
		ctx.Reply("!DWALL: SHOW EVERYBODY'S RANK, EVEN THE WORTHLESS FOLK")
	case "DONGWINNERS":
		ctx.Reply("!DONGWINNERS [NUMBER]: SHOW THE PEOPLE WHO FUCKING MATTER, BY DEFAULT DOES THE TOP 2 FOR THE DAY... JUST LIKE A SLOW NIGHT FOR YERMOM")
	case "DONGRANKME":
		ctx.Reply("!DONGRANKME: SHOW YOUR RELATIVE WORTH")
	case "SIZE":
		ctx.Reply("!SIZE [USERNAME]: GIVES YOU THE ONLY THING THAT MATTERS ABOUT SOMEBODY: SIZE")
	case "SIZEME":
		ctx.Reply("!SIZEME: TELLS YOU IF YOU ARE WORTH ANYTHING TO SOCIETY")
	case "HELP":
		ctx.Reply("OH WOW YOU ARE SO META I AM SO IMPRESSED WE SHOULD GO HAVE SEX NOW")
	case "DONGME":
		ctx.Reply("!DONGME: SHOWS HOW MUCH OF A MAN YOU ARE")
	case "REDONGME":
		ctx.Reply("!REDONGME: LETS YOU TRY TO MAKE YOURSELF INTO MORE OF A MAN BUT WITH DANGER RISK!")
	case "OMAKASE":
		ctx.Reply("!OMAKASE [TOOLNAME]: MAKES TOOLS REALLY GREAT INSTEAD OF GIANT PILES OF POORLY-ARCHITECTED BULLSHIT")
	case "UNDONG":
		ctx.Reply("!UNDONG <USERNAME>: ADMIN ONLY - REMOVES A USER'S DONG FOR TODAY")
	default:
		ctx.Reply("!" + cmd + ": DOES SOMETHING AWESOME")
	}
}
