package poker

import (
	"fmt"
	"log"
	"math/bits"
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

func evalScore5(v int, a, b, c, d, e int) eval {
	return eval{rank: v*16*16*16*16*16 + a*16*16*16*16 + b*16*16*16 + c*16*16 + d*16 + e}
}

func isFlush(c []Card) bool {
	if len(c) != 5 {
		return false
	}
	and := (1 << (c[0] & 3)) & (1 << (c[1] & 3)) & (1 << (c[2] & 3)) & (1 << (c[3] & 3)) & (1 << (c[4] & 3))
	return and != 0
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

// DescribeShort describes a 3, 5 or 7 card poker hand with enough detail
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

func poptop(x uint16) (int, uint16) {
	lz := bits.LeadingZeros16(x)
	if lz == 16 {
		return 0, 0
	}
	return 17 - lz, x &^ (1 << (15 - lz))
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
// It's slow, but a little bit optimized so that the table construction
// is relatively fast.
func evalSlow(c []Card, replace, text bool) (eval, error) {
	if len(c) == 7 {
		return evalSlow7(c, replace, text)
	}
	flush := isFlush(c)
	ranks := [13]int{}
	dupes := [6]int{}  // uniqs, pairs, trips, quads, quins
	str8s := [13]int{} // finds straights
	str8top := 0       // set to the top card of the straight, if any
	var rankBits [6]uint16
	for _, ci := range c {
		cr := (int(ci>>2) & 15) + 1
		rawr := (cr + 11) % 13
		rankBits[ranks[rawr]] |= 1 << rawr
		ranks[rawr]++
		dupes[ranks[rawr]]++
		dupes[ranks[rawr]-1]--
		for i := 0; i < 5; i++ {
			idx := (cr + i) % 13
			str8s[idx] |= 1 << uint(i)
			// Make sure to exclude wrap-around straights headed by 2, 3, 4.
			if str8s[idx] == 31 && (idx <= 1 || idx >= 5) {
				str8top = (idx+12)%13 + 1
			}
		}
	}
	rankBits[0] &^= rankBits[1]
	rankBits[1] &^= rankBits[2]
	rankBits[2] &^= rankBits[3]
	rankBits[3] &^= rankBits[4]
	rankBits[4] &^= rankBits[5]
	if !flush && str8top == 0 && dupes[1] == len(c) { // No pair
		var a, b, c, d, e int
		a, rankBits[0] = poptop(rankBits[0])
		b, rankBits[0] = poptop(rankBits[0])
		c, rankBits[0] = poptop(rankBits[0])
		d, rankBits[0] = poptop(rankBits[0])
		e, rankBits[0] = poptop(rankBits[0])
		if text {
			return evalScore("%s-%s-%s-%s-%s", 0, a, b, c, d, e), nil
		}
		return evalScore5(0, a, b, c, d, e), nil
	}
	if dupes[2] == 1 && dupes[3] == 0 { // One pair
		var p, a, b, c int
		p, rankBits[1] = poptop(rankBits[1])
		a, rankBits[0] = poptop(rankBits[0])
		b, rankBits[0] = poptop(rankBits[0])
		c, rankBits[0] = poptop(rankBits[0])
		if text {
			return evalScore("%[1]s%[1]s-%s-%s-%s", 1, p, a, b, c), nil
		}
		return evalScore5(1, p, a, b, c, 0), nil
	}
	if dupes[2] == 2 { // Two pair
		var p, q, a int
		p, rankBits[1] = poptop(rankBits[1])
		q, rankBits[1] = poptop(rankBits[1])
		a, rankBits[0] = poptop(rankBits[0])
		if text {
			return evalScore("%[1]s%[1]s-%[2]s%[2]s-%[3]s", 2, p, q, a), nil
		}
		return evalScore5(2, p, q, a, 0, 0), nil
	}
	if dupes[3] == 1 && dupes[2] == 0 { // Trips
		if replace {
			var t, a, b int
			a, rankBits[0] = poptop(rankBits[0])
			b, rankBits[0] = poptop(rankBits[0])
			t, rankBits[2] = poptop(rankBits[2])
			if text {
				return evalScore("%[1]s%[1]s%[1]s-%s-%s", 3, t, a, b), nil
			}
			return evalScore5(3, t, a, b, 0, 0), nil
		}
		if len(c) == 5 {
			var t int
			t, rankBits[2] = poptop(rankBits[2])
			if text {
				return evalScore("%[1]s%[1]s%[1]s-x-y", 3, t), nil // ignore kickers
			}
			return evalScore5(3, t, 0, 0, 0, 0), nil
		}
		var t int
		t, rankBits[2] = poptop(rankBits[2])
		if text {
			return evalScore("%[1]s%[1]s%[1]s", 3, t), nil
		}
		return evalScore5(3, t, 0, 0, 0, 0), nil
	}
	if str8top != 0 && !flush { // Straight
		if text {
			return evalScore("%s straight", 4, (str8top+11)%13+2), nil
		}
		return evalScore5(4, (str8top+11)%13+2, 0, 0, 0, 0), nil
	}
	if flush && str8top == 0 { // Flush
		var a, b, c, d, e int
		a, rankBits[0] = poptop(rankBits[0])
		b, rankBits[0] = poptop(rankBits[0])
		c, rankBits[0] = poptop(rankBits[0])
		d, rankBits[0] = poptop(rankBits[0])
		e, rankBits[0] = poptop(rankBits[0])
		if text {
			return evalScore("%s%s%s%s%s flush", 5, a, b, c, d, e), nil
		}
		return evalScore5(5, a, b, c, d, e), nil
	}
	if dupes[2] == 1 && dupes[3] == 1 { // Full house
		var t, p int
		t, rankBits[2] = poptop(rankBits[2])
		p, rankBits[1] = poptop(rankBits[1])
		if replace {
			if text {
				return evalScore("%[1]s%[1]s%[1]s-%[2]s%[2]s", 6, t, p), nil
			}
			return evalScore5(6, t, p, 0, 0, 0), nil
		}
		if text {
			return evalScore("%[1]s%[1]s%[1]s-xx", 6, t), nil // ignore lower pair
		}
		return evalScore5(6, t, 0, 0, 0, 0), nil // ignore lower pair
	}
	if dupes[4] == 1 { // Quads
		var q, a int
		q, rankBits[3] = poptop(rankBits[3])
		a, rankBits[0] = poptop(rankBits[0])
		if replace {
			if text {
				return evalScore("%[1]s%[1]s%[1]s%[1]s-%[2]s", 7, q, a), nil
			}
			return evalScore5(7, q, a, 0, 0, 0), nil
		}
		if text {
			return evalScore("%[1]s%[1]s%[1]s%[1]s-x", 7, q), nil // ignore kicker
		}
		return evalScore5(7, q, 0, 0, 0, 0), nil
	}
	if str8top != 0 && flush { // Straight flush
		if text {
			return evalScore("%s straight flush", 8, (str8top+11)%13+2), nil
		}
		return evalScore5(8, (str8top+11)%13+2, 0, 0, 0, 0), nil
	}
	if dupes[5] == 1 { // 5-kind
		var q int
		q, rankBits[4] = poptop(rankBits[4])
		if text {
			return evalScore("%[1]s%[1]s%[1]s%[1]s%[1]s", 9, q), nil
		}
		return evalScore5(9, q, 0, 0, 0, 0), nil
	}
	return eval{}, fmt.Errorf("failed to eval hand %v", c)
}

// ScoreMax is the largest possible rank of hand returned by the Eval
// functions
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

// EvalSlow takes a 3-, 5- or 7- card poker hand and returns a number
// which can be used to rank it against other poker hands.
// The returned value is in the range 0 to ScoreMax.
// This function should not generally be used, and Eval3, Eval5 or Eval7
// used instead. It uses a straightforward algorithm for hand-ranking.
func EvalSlow(c []Card) int16 {
	ev, _ := evalSlow(c, true, false)
	return evalInfo.slowRankToPacked[ev.rank]
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

func mustMakeCard(s Suit, r Rank) Card {
	c, err := MakeCard(s, r)
	if err != nil {
		panic(err)
	}
	return c
}

func makeEvalInfo() *evalInfos {
	ei := &evalInfos{}
	// Enumerate enough 3-card hands to categorize all hands.
	// There's no flushes, so suits are not important except
	// that there can't be duplicate cards.
	hand3 := map[int][]Card{}
	for a := 0; a < 13; a++ {
		carda := mustMakeCard(Club, Rank(a+1))
		for b := a; b < 13; b++ {
			cardb := mustMakeCard(Diamond, Rank(b+1))
			for c := b; c < 13; c++ {
				cardc := mustMakeCard(Heart, Rank(c+1))
				h3 := []Card{carda, cardb, cardc}
				ev, err := evalSlow(h3, true, false)
				if err != nil {
					panic(err)
				}
				if _, ok := hand3[ev.rank]; !ok {
					hand3[ev.rank] = h3
				}
			}
		}
	}
	hand5 := map[int][]Card{}

	// Enumerate all 5-card flush hands.
	for a := 0; a < 13; a++ {
		carda := mustMakeCard(Club, Rank(a+1))
		for b := a + 1; b < 13; b++ {
			cardb := mustMakeCard(Club, Rank(b+1))
			for c := b + 1; c < 13; c++ {
				cardc := mustMakeCard(Club, Rank(c+1))
				for d := c + 1; d < 13; d++ {
					cardd := mustMakeCard(Club, Rank(d+1))
					for e := d + 1; e < 13; e++ {
						carde := mustMakeCard(Club, Rank(e+1))
						h5 := []Card{carda, cardb, cardc, cardd, carde}
						ev, err := evalSlow(h5, true, false)
						if err != nil {
							panic(err)
						}
						if _, ok := hand5[ev.rank]; !ok {
							hand5[ev.rank] = h5
						}
					}
				}
			}
		}
	}

	// Enumerate all 5-card non-flush hands.
	for a := 0; a < 13; a++ {
		carda := mustMakeCard(Club, Rank(a+1))
		for b := a; b < 13; b++ {
			cardb := mustMakeCard(Diamond, Rank(b+1))
			for c := b; c < 13; c++ {
				cardc := mustMakeCard(Heart, Rank(c+1))
				for d := c; d < 13; d++ {
					cardd := mustMakeCard(Spade, Rank(d+1))
					for e := d; e < 13; e++ {
						carde := mustMakeCard(Club, Rank(e+1))
						h5 := []Card{carda, cardb, cardc, cardd, carde}
						ev, err := evalSlow(h5, true, false)
						if err != nil {
							panic(err)
						}
						if _, ok := hand5[ev.rank]; !ok {
							hand5[ev.rank] = h5
						}
					}
				}
			}
		}
	}

	// Aggregate and pack scores.
	allScores := []int{}
	for k := range hand3 {
		allScores = append(allScores, k)
	}
	for k := range hand5 {
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
