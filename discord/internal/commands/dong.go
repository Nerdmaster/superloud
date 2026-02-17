package commands

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/dialmaster/superloud-discord/internal/util"
)

var places = []string{"FIRST", "SECOND", "THIRD", "FOURTH", "FIFTH", "SIXTH", "SEVENTH", "EIGHTH", "NINTH", "TENTH"}

// rollPenalty returns the penalty for the given number of rerolls.
func rollPenalty(count int) int {
	vals := []int{0, -1, -3, -6, -9, -12, -18, -24, -36, -48, -60, -72, -84, -96, -108}
	if count < len(vals) {
		return vals[count]
	}
	return vals[len(vals)-1]
}

// fairDongSize returns a value in 1/2cm units of a randomized, normalized dong length.
func fairDongSize(rng *rand.Rand) int {
	average := 28
	roll := util.DiceRoll(3, 15, rng) - 22

	var percent float64
	if roll > 0 {
		percent = float64(100+roll*6) / 100.0
	} else {
		percent = float64(100+roll*3) / 100.0
	}

	return int(percent * float64(average))
}

// microdongSize returns a value in 1/2cm units for Micropenis Monday.
func microdongSize(rng *rand.Rand) int {
	average := 28
	roll := util.DiceRoll(3, 5, rng) - 22
	percent := float64(100+roll*3) / 100.0
	return int(percent * float64(average))
}

// computeSize computes dong size for a user, storing the result.
func (r *Registry) computeSize(userHash int64, userName string) int {
	mulligans := r.Redongs[userHash]
	sizeModifier := rollPenalty(mulligans)

	// Create seeded RNG from user hash + date + mulligan offset
	dateInt := dateToInt(time.Now())
	seed := userHash + int64(dateInt) + int64(mulligans*53)
	rng := rand.New(rand.NewSource(seed))

	var size int
	if time.Now().Weekday() == time.Monday {
		size = max(2, microdongSize(rng)+sizeModifier)
	} else {
		size = max(2, fairDongSize(rng)+sizeModifier)
	}

	r.SizeData[userHash] = &SizeEntry{
		Size: size,
		Nick: userName,
		Hash: userHash,
	}

	return size
}

func dateToInt(t time.Time) int {
	return t.Year()*10000 + int(t.Month())*100 + t.Day()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (r *Registry) cmdDongMe(ctx *CommandContext) {
	r.sendDong(ctx)
}

func (r *Registry) cmdReDongMe(ctx *CommandContext) {
	userHash := util.UserHash(r.Filters.ResolveAlias(ctx.UserID))

	// Clear old size data
	delete(r.SizeData, userHash)

	// Increment mulligan count
	r.Redongs[userHash]++

	r.sendDong(ctx)
}

func (r *Registry) sendDong(ctx *CommandContext) {
	userHash := util.UserHash(r.Filters.ResolveAlias(ctx.UserID))
	size := r.computeSize(userHash, ctx.UserName)
	ctx.Reply(fmt.Sprintf("8%sD", strings.Repeat("=", size)))
}

func (r *Registry) cmdSizeMe(ctx *CommandContext) {
	ctx.Params = []string{ctx.UserName}
	r.cmdSize(ctx)
}

func (r *Registry) cmdSize(ctx *CommandContext) {
	if len(ctx.Params) == 0 {
		ctx.Params = []string{"SIZE"}
		r.cmdHelp(ctx)
		return
	}

	name := strings.ToUpper(ctx.Params[0])
	// Strip Discord mention formatting
	name = strings.TrimPrefix(name, "<@")
	name = strings.TrimPrefix(name, "!")
	name = strings.TrimSuffix(name, ">")

	var msg string
	for _, entry := range r.SizeData {
		if strings.ToUpper(entry.Nick) == name {
			cm := float64(entry.Size) / 2.0
			inches := cm / 2.54
			if name == strings.ToUpper(ctx.UserName) {
				msg = fmt.Sprintf("HEY %s YOUR DONG IS %0.1f INCHES (%0.1f CM)", name, inches, cm)
			} else {
				msg = fmt.Sprintf("%s'S DONG IS %0.1f INCHES (%0.1f CM)", name, inches, cm)
			}
			break
		}
	}

	if msg == "" {
		msg = fmt.Sprintf("ONOES THERE IS NO DONG FOR %s", name)
	}

	ctx.Reply(strings.ToUpper(msg))
}

func (r *Registry) cmdBiggestDong(ctx *CommandContext) {
	if len(r.SizeData) == 0 {
		ctx.Reply("ONOES NO DONGS TODAY SIRS")
		return
	}

	ranked := r.rankBySize()
	winners := ranked[0]
	names := make([]string, len(winners))
	for i, w := range winners {
		names[i] = w.Nick
	}
	nameText := strings.ToUpper(userlistText(names))
	size := winners[0].Size

	cm := float64(size) / 2.0
	inches := cm / 2.54

	if len(winners) <= 1 {
		ctx.Reply(fmt.Sprintf("THE BIGGEST I'VE SEEN TODAY IS %s'S WHICH IS %0.1f INCHES (%0.1f CM)", nameText, inches, cm))
	} else {
		word := "BOTH"
		if len(winners) > 2 {
			word = "ALL"
		}
		ctx.Reply(fmt.Sprintf("THE BIGGEST I'VE SEEN TODAY IS... OMFG... IT'S A TIE BETWEEN %s!  THEY'RE %s %0.1f INCHES (%0.1f CM)!!!1!!", nameText, word, inches, cm))
	}
}

func (r *Registry) cmdDongWinners(ctx *CommandContext) {
	ranked := r.rankBySize()
	if len(ranked) == 0 {
		ctx.Reply("ONOES NO DONGS TODAY SIRS")
		return
	}

	numPlaces := 2
	if len(ctx.Params) > 0 {
		fmt.Sscanf(ctx.Params[0], "%d", &numPlaces)
	}
	if numPlaces > 10 || numPlaces < 2 {
		numPlaces = 2
	}

	if numPlaces > len(ranked) {
		numPlaces = len(ranked)
	}

	var output []string
	for i := 0; i < numPlaces; i++ {
		winners := ranked[i]
		names := make([]string, len(winners))
		for j, w := range winners {
			names[j] = w.Nick
		}
		nameText := strings.ToUpper(userlistText(names))
		output = append(output, fmt.Sprintf("IN %s PLACE WE HAVE %s", places[i], nameText))
	}

	ctx.Reply(strings.Join(output, "; "))
}

func (r *Registry) cmdDongRankMe(ctx *CommandContext) {
	userHash := util.UserHash(r.Filters.ResolveAlias(ctx.UserID))
	entry := r.SizeData[userHash]
	if entry == nil {
		ctx.Reply("YOU DON'T HAVE A DONG DUMBASS")
		return
	}

	rank := r.sizeToRank(entry.Size)
	ctx.Reply(fmt.Sprintf("YOU ARE CURRENTLY RANKED %s!", placeText(rank)))
}

func (r *Registry) cmdDWall(ctx *CommandContext) {
	ctx.Params = []string{"10"}
	r.cmdDongWinners(ctx)
}

// rankBySize returns users grouped by size, sorted descending.
func (r *Registry) rankBySize() [][]*SizeEntry {
	bySize := make(map[int][]*SizeEntry)
	for _, entry := range r.SizeData {
		bySize[entry.Size] = append(bySize[entry.Size], entry)
	}

	sizes := make([]int, 0, len(bySize))
	for s := range bySize {
		sizes = append(sizes, s)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(sizes)))

	result := make([][]*SizeEntry, len(sizes))
	for i, s := range sizes {
		result[i] = bySize[s]
	}
	return result
}

func (r *Registry) sizeToRank(size int) int {
	bySize := make(map[int]bool)
	for _, entry := range r.SizeData {
		bySize[entry.Size] = true
	}

	sizes := make([]int, 0, len(bySize))
	for s := range bySize {
		sizes = append(sizes, s)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(sizes)))

	for rank, s := range sizes {
		if s == size {
			return rank + 1
		}
	}
	return len(sizes)
}

func placeText(place int) string {
	if place >= 1 && place <= len(places) {
		return places[place-1]
	}
	return fmt.Sprintf("NUMBER %d", place)
}

func userlistText(userlist []string) string {
	sorted := make([]string, len(userlist))
	copy(sorted, userlist)
	sort.Strings(sorted)

	switch len(sorted) {
	case 1:
		return sorted[0]
	case 2:
		return sorted[0] + " AND " + sorted[1]
	default:
		return strings.Join(sorted[:len(sorted)-1], ", ") + ", AND " + sorted[len(sorted)-1]
	}
}
