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
		ranks[rawr] += 1
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
		} else {
			return evalScore5(0, a, b, c, d, e), nil
		}
	}
	if dupes[2] == 1 && dupes[3] == 0 { // One pair
		var p, a, b, c int
		p, rankBits[1] = poptop(rankBits[1])
		a, rankBits[0] = poptop(rankBits[0])
		b, rankBits[0] = poptop(rankBits[0])
		c, rankBits[0] = poptop(rankBits[0])
		if text {
			return evalScore("%[1]s%[1]s-%s-%s-%s", 1, p, a, b, c), nil
		} else {
			return evalScore5(1, p, a, b, c, 0), nil
		}
	}
	if dupes[2] == 2 { // Two pair
		var p, q, a int
		p, rankBits[1] = poptop(rankBits[1])
		q, rankBits[1] = poptop(rankBits[1])
		a, rankBits[0] = poptop(rankBits[0])
		if text {
			return evalScore("%[1]s%[1]s-%[2]s%[2]s-%[3]s", 2, p, q, a), nil
		} else {
			return evalScore5(2, p, q, a, 0, 0), nil
		}
	}
	if dupes[3] == 1 && dupes[2] == 0 { // Trips
		if replace {
			var t, a, b int
			a, rankBits[0] = poptop(rankBits[0])
			b, rankBits[0] = poptop(rankBits[0])
			t, rankBits[2] = poptop(rankBits[2])
			if text {
				return evalScore("%[1]s%[1]s%[1]s-%s-%s", 3, t, a, b), nil
			} else {
				return evalScore5(3, t, a, b, 0, 0), nil
			}
		}
		if len(c) == 5 {
			var t int
			t, rankBits[2] = poptop(rankBits[2])
			if text {
				return evalScore("%[1]s%[1]s%[1]s-x-y", 3, t), nil // ignore kickers
			} else {
				return evalScore5(3, t, 0, 0, 0, 0), nil
			}
		}
		var t int
		t, rankBits[2] = poptop(rankBits[2])
		if text {
			return evalScore("%[1]s%[1]s%[1]s", 3, t), nil
		} else {
			return evalScore5(3, t, 0, 0, 0, 0), nil
		}
	}
	if str8top != 0 && !flush { // Straight
		if text {
			return evalScore("%s straight", 4, (str8top+11)%13+2), nil
		} else {
			return evalScore5(4, (str8top+11)%13+2, 0, 0, 0, 0), nil
		}
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
		} else {
			return evalScore5(5, a, b, c, d, e), nil
		}
	}
	if dupes[2] == 1 && dupes[3] == 1 { // Full house
		var t, p int
		t, rankBits[2] = poptop(rankBits[2])
		p, rankBits[1] = poptop(rankBits[1])
		if replace {
			if text {
				return evalScore("%[1]s%[1]s%[1]s-%[2]s%[2]s", 6, t, p), nil
			} else {
				return evalScore5(6, t, p, 0, 0, 0), nil
			}
		}
		if text {
			return evalScore("%[1]s%[1]s%[1]s-xx", 6, t), nil // ignore lower pair
		} else {
			return evalScore5(6, t, 0, 0, 0, 0), nil // ignore lower pair
		}
	}
	if dupes[4] == 1 { // Quads
		var q, a int
		q, rankBits[3] = poptop(rankBits[3])
		a, rankBits[0] = poptop(rankBits[0])
		if replace {
			if text {
				return evalScore("%[1]s%[1]s%[1]s%[1]s-%[2]s", 7, q, a), nil
			} else {
				return evalScore5(7, q, a, 0, 0, 0), nil
			}
		}
		if text {
			return evalScore("%[1]s%[1]s%[1]s%[1]s-x", 7, q), nil // ignore kicker
		} else {
			return evalScore5(7, q, 0, 0, 0, 0), nil
		}
	}
	if str8top != 0 && flush { // Straight flush
		if text {
			return evalScore("%s straight flush", 8, (str8top+11)%13+2), nil
		} else {
			return evalScore5(8, (str8top+11)%13+2, 0, 0, 0, 0), nil
		}
	}
	if dupes[5] == 1 { // 5-kind
		var q int
		q, rankBits[4] = poptop(rankBits[4])
		if text {
			return evalScore("%[1]s%[1]s%[1]s%[1]s%[1]s", 9, q), nil
		} else {
			return evalScore5(9, q, 0, 0, 0, 0), nil
		}
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

// EvalSlow takes a 3-, 5- or 7- card poker hand and returns a number
// which can be used to rank it against other poker hands.
// The returned value is in the range 0 to ScoreMax.
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
