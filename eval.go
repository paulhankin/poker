package poker

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

type eval struct {
	desc string
	rank int
}

// evalScore condenses the strength of a hand given its type
// and cards into a single int.
func evalScore(f string, v int, c ...int) eval {
	r := v
	for i := 0; i < 5; i++ {
		r *= 16
		if i < len(c) {
			r += c[i]
		}
	}
	args := make([]interface{}, len(c))
	for i := range c {
		if c[i] == 0 {
			args[i] = ""
			continue
		}
		args[i] = Rank((c[i]-1)%13 + 1)
	}
	return eval{
		desc: fmt.Sprintf(f, args...),
		rank: r,
	}
}

func evalScoreNoText(f string, v int, c ...int) eval {
	r := v
	for i := 0; i < 5; i++ {
		r *= 16
		if i < len(c) {
			r += c[i]
		}
	}
	return eval{
		rank: r,
	}
}

// find picks the nth highest rank of r which is equal to k,
// returning a number which is higher for higher cards.
// Returns 0 if there is none.
func find(k int, r *[13]int, n int) int {
	for i := 12; i >= 0; i-- {
		if r[i] == k {
			if n == 0 {
				return i + 2
			}
			n--
		}
	}
	return 0
}

func find1(r *[13]int, n int) int {
	return find(1, r, n)
}

func find2(r *[13]int, n int) int {
	return find(2, r, n)
}

func isFlush(c []Card) bool {
	if len(c) != 5 {
		return false
	}
	for _, ci := range c {
		if ci.Suit() != c[0].Suit() {
			return false
		}
	}
	return true
}

// Describe fully describes a 3, 5 or 7 card poker hand.
func Describe(c []Card) (string, error) {
	eval, err := evalSlow(c, true, true)
	if err != nil {
		return "", err
	}
	// The description of a three-card hand includes trailing dashes
	// where kickers are missing. We remove them.
	return strings.TrimRight(eval.desc, "-"), nil
}

// Describe describes a 3, 5 or 7 card poker hand with enough detail
// to compare it to another poker hand which shares no cards in common.
// For example, KKK-87 is represented as KKK-x-y since the kickers can
// never matter (except that they are different).
func DescribeShort(c []Card) (string, error) {
	eval, err := evalSlow(c, false, true)
	if err != nil {
		return "", err
	}
	// The description of a three-card hand includes trailing dashes
	// where kickers are missing. We remove them.
	return strings.TrimRight(eval.desc, "-"), nil
}

func evalSlow7(c []Card, replace, text bool) (eval, error) {
	idx := [5]int{4, 3, 2, 1, 0}
	var bestEval eval
	var bestHand [5]Card
	for {
		h := [5]Card{c[idx[0]], c[idx[1]], c[idx[2]], c[idx[3]], c[idx[4]]}
		ev, err := evalSlow(h[:], replace, false)
		if err != nil {
			return eval{}, err
		}
		if ev.rank > bestEval.rank {
			bestEval = ev
			bestHand = h
		}
		if idx[0] < 6 {
			idx[0]++
		} else if idx[1] < 5 {
			idx[1]++
			idx[0] = idx[1] + 1
		} else if idx[2] < 4 {
			idx[2]++
			idx[1] = idx[2] + 1
			idx[0] = idx[1] + 1
		} else if idx[3] < 3 {
			idx[3]++
			idx[2] = idx[3] + 1
			idx[1] = idx[2] + 1
			idx[0] = idx[1] + 1
		} else if idx[4] < 2 {
			idx[4]++
			idx[3] = idx[4] + 1
			idx[2] = idx[3] + 1
			idx[1] = idx[2] + 1
			idx[0] = idx[1] + 1
		} else {
			var err error
			if text {
				bestEval, err = evalSlow(bestHand[:], replace, true)
			}
			return bestEval, err
		}
	}
}

// evalSlow evaluates a 3- or 5- card poker hand.
// The result is a number which can be compared
// with other hand's evaluations to correctly rank them as poker
// hands.
// If "replace" is false, then details are dropped of hands that can't be
// used for comparing against hands which are drawn from the same
// deck (for example: the kickers with trip aces don't matter).
//
// This function is used to build tables for fast hand evaluation.
func evalSlow(c []Card, replace, text bool) (eval, error) {
	if len(c) == 7 {
		return evalSlow7(c, replace, text)
	}
	es := evalScore
	if !text {
		es = evalScoreNoText
	}
	flush := isFlush(c)
	ranks := &[13]int{}
	dupes := [6]int{}  // uniqs, pairs, trips, quads, quins
	str8s := [13]int{} // finds straights
	str8top := 0       // set to the top card of the straight, if any
	for _, ci := range c {
		rawr := ci.RawRank()
		ranks[rawr] += 1
		dupes[ranks[rawr]]++
		dupes[ranks[rawr]-1]--
		for i := 0; i < 5; i++ {
			idx := (int(ci.Rank()) + i) % 13
			str8s[idx] |= 1 << uint(i)
			// Make sure to exclude wrap-around straights headed by 2, 3, 4.
			if str8s[idx] == 31 && (idx <= 1 || idx >= 5) {
				str8top = (idx+12)%13 + 1
			}
		}
	}
	if !flush && str8top == 0 && dupes[1] == len(c) { // No pair
		return es("%s-%s-%s-%s-%s", 0, find1(ranks, 0), find1(ranks, 1), find1(ranks, 2), find1(ranks, 3), find1(ranks, 4)), nil
	}
	if dupes[2] == 1 && dupes[3] == 0 { // One pair
		return es("%[1]s%[1]s-%s-%s-%s", 1, find2(ranks, 0), find1(ranks, 0), find1(ranks, 1), find1(ranks, 2)), nil
	}
	if dupes[2] == 2 { // Two pair
		return es("%[1]s%[1]s-%[2]s%[2]s-%[3]s", 2, find2(ranks, 0), find2(ranks, 1), find1(ranks, 0)), nil
	}
	if dupes[3] == 1 && dupes[2] == 0 { // Trips
		if replace {
			return es("%[1]s%[1]s%[1]s-%s-%s", 3, find(3, ranks, 0), find1(ranks, 0), find1(ranks, 1)), nil
		}
		if len(c) == 5 {
			return es("%[1]s%[1]s%[1]s-x-y", 3, find(3, ranks, 0)), nil // ignore kickers
		}
		return es("%[1]s%[1]s%[1]s", 3, find(3, ranks, 0)), nil
	}
	if str8top != 0 && !flush { // Straight
		return es("%s straight", 4, (str8top+11)%13+2), nil
	}
	if flush && str8top == 0 { // Flush
		return es("%s%s%s%s%s flush", 5, find1(ranks, 0), find1(ranks, 1), find1(ranks, 2), find1(ranks, 3), find1(ranks, 4)), nil
	}
	if dupes[2] == 1 && dupes[3] == 1 { // Full house
		if replace {
			return es("%[1]s%[1]s%[1]s-%[2]s%[2]s", 6, find(3, ranks, 0), find2(ranks, 0)), nil
		}
		return es("%[1]s%[1]s%[1]s-xx", 6, find(3, ranks, 0)), nil // ignore lower pair
	}
	if dupes[4] == 1 { // Quads
		if replace {
			return es("%[1]s%[1]s%[1]s%[1]s-%[2]s", 7, find(4, ranks, 0), find1(ranks, 0)), nil
		}
		return es("%[1]s%[1]s%[1]s%[1]s-x", 7, find(4, ranks, 0)), nil // ignore kicker
	}
	if str8top != 0 && flush { // Straight flush
		return es("%s straight flush", 8, (str8top+11)%13+2), nil
	}
	if dupes[5] == 1 { // 5-kind
		return es("%[1]s%[1]s%[1]s%[1]s%[1]s", 9, find(5, ranks, 0)), nil
	}
	return eval{}, fmt.Errorf("failed to eval hand %v", c)
}

// ScoreMax is the largest possible result from Eval (with replace=true).
const ScoreMax = 7929

type evalInfos struct {
	rankTo5          [ScoreMax + 1][]Card
	rankTo3          [ScoreMax + 1][]Card
	slowRankToPacked map[int]int16
}

var evalInfo *evalInfos = makeEvalInfo()

// EvalToHand5 returns an example 5-card hand with the given
// eval score. The xsecond return value is whether the result is valid.
func EvalToHand5(e int16) ([]Card, bool) {
	if e < 0 || e > ScoreMax {
		return nil, false
	}
	return evalInfo.rankTo5[e], len(evalInfo.rankTo5[e]) != 0
}

// EvalToHand3 returns an example 3-card hand with the given
// eval score. The second return value is whether the result is valid.
func EvalToHand3(e int16) ([]Card, bool) {
	if e < 0 || e > ScoreMax {
		return nil, false
	}
	return evalInfo.rankTo3[e], len(evalInfo.rankTo3[e]) != 0
}

// Eval takes a 3- or 5- card poker hand and returns a number
// which can be used to rank it against other poker hands.
// The returned value is in the range 0 to ScoreMax.
func Eval(c []Card) int16 {
	ev, _ := evalSlow(c, true, false)
	return evalInfo.slowRankToPacked[ev.rank]
}

func eval5idx(c *[7]Card, idx [5]int) int16 {
	h := [5]Card{c[idx[0]], c[idx[1]], c[idx[2]], c[idx[3]], c[idx[4]]}
	return Eval(h[:])
}

// Eval7 returns the ranking of the best 5-card hand
// that's a subset of the given 7 cards.
func Eval7(c *[7]Card) int16 {
	idx := [5]int{4, 3, 2, 1, 0}
	var best int16
	for {
		if ev := eval5idx(c, idx); ev > best {
			best = ev
		}
		if idx[0] < 6 {
			idx[0]++
		} else if idx[1] < 5 {
			idx[1]++
			idx[0] = idx[1] + 1
		} else if idx[2] < 4 {
			idx[2]++
			idx[1] = idx[2] + 1
			idx[0] = idx[1] + 1
		} else if idx[3] < 3 {
			idx[3]++
			idx[2] = idx[3] + 1
			idx[1] = idx[2] + 1
			idx[0] = idx[1] + 1
		} else if idx[4] < 2 {
			idx[4]++
			idx[3] = idx[4] + 1
			idx[2] = idx[3] + 1
			idx[1] = idx[2] + 1
			idx[0] = idx[1] + 1
		} else {
			return best
		}
	}
}

// Eval5 is an optimized version of Eval which requires a 5-card hand.
func Eval5(c *[5]Card) int16 {
	return Eval(c[:])
}

// Eval3 is an optimized version of Eval which requires a 3-card hand.
func Eval3(c *[3]Card) int16 {
	return Eval(c[:])
}

func nextIdx(ix []int, k int, dupes int) bool {
	i := 0
	for {
		ix[i]++
		if i+1 == len(ix) || ix[i] != ix[i+1]+dupes {
			return ix[i] < k
		}
		ix[i] = i * (1 - dupes)
		i++
	}
}

func makeEvalInfo() *evalInfos {
	ei := &evalInfos{}

	uniqScores := map[int]bool{}
	hand5, hand3 := map[int][]Card{}, map[int][]Card{}
	for _, size := range []int{3, 5} {
		indexes := make([]int, size)
		hand := make([]Card, size)
		// We iterate over enough hands to categorize _all_ hands.
		// For non-flush hands we allow duplicate cards (eg: pairs)
		// but fix the suits. For flush hands, we don't allow duplicate
		// cards, and fix the suit to be spades.
		s := []Suit{Club, Diamond, Heart, Spade, Club}
		flushTop := size / 5 // 0 if size=3, 1 if size=5.
		for flush := 0; flush <= flushTop; flush++ {
			if flush == 1 {
				for i := range indexes {
					indexes[i] = i
				}
			}
			for {
				for i, ix := range indexes {
					suit := Spade
					if flush == 0 {
						suit = s[i]
					}
					var err error
					hand[i], err = MakeCard(suit, Rank(ix+1))
					if err != nil {
						log.Fatalf("failed to create card: %s", err)
					}
				}
				ev, err := evalSlow(hand, true, false)
				if err != nil {
					log.Fatalf("evalSlow(%v) gave error %s", hand, err)
				}
				if size == 3 {
					hand3[ev.rank] = append([]Card{}, hand...)
				} else {
					hand5[ev.rank] = append([]Card{}, hand...)
				}
				uniqScores[ev.rank] = true
				if !nextIdx(indexes, 13, 1-flush) {
					break
				}
			}
		}
	}
	// Aggregate and pack scores.
	allScores := []int{}
	for k := range uniqScores {
		allScores = append(allScores, k)
	}
	sort.Ints(allScores)

	ei.slowRankToPacked = map[int]int16{}
	for i, k := range allScores {
		ei.slowRankToPacked[k] = int16(i)
	}

	for rank, packedRank := range ei.slowRankToPacked {
		ei.rankTo5[packedRank] = hand5[rank]
		ei.rankTo3[packedRank] = hand3[rank]
	}
	if ScoreMax != len(allScores)-1 {
		log.Fatalf("Expected max score of %d, but found %d", ScoreMax, len(allScores)-1)
	}
	return ei
}
