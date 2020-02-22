package poker

import (
	"sort"
	"strings"
)

// Hand64 is a hand with up to N<=7 cards in it, stored
// in the lowest N bytes.
type Hand64 uint64

// Hand64Canonical is a compressed version of Hand64.
// It is like a Hand64 except that the cards are sorted, and the suits are
// canonicalized to appear in increasing order, so that the first card
// is always a club, the second a club
// or diamond, the third a club, diamond, or heart (only if
// the second card was a diamond), and so on.
// If there's no way to extend the hand to 7 cards so that
// a particular suit can form a flush, then that suit will
// be coalesced with all other suits that can't form flushes.
type Hand64Canonical uint64

// CardsN returns the first n cards in the hand.
// If n is more than 7, only the first 7 cards are returned.
func (h Hand64) CardsN(n int) []Card {
	if n <= 0 {
		return nil
	}
	if n > 7 {
		n = 7
	}
	c := make([]Card, n)
	for i := 0; i < n; i++ {
		c[i] = h.Card(i)
	}
	return c
}

// Add adds a card to an n-card hand. The result
// is a hand with no x-suit cards.
// The card can't be added if:
//  - it's already in the hand
//  - it would result in 5 cards of the same rank in the hand
func (h Hand64Canonical) Add(n int, c Card) (Hand64, bool) {
	rc := 0
	for i := 0; i < n; i++ {
		ci := Hand64(h).Card(i)
		if ci == c {
			return 0, false
		}
		if ci.Rank() == c.Rank() {
			rc++
		}
	}
	if rc >= 4 {
		return 0, false
	}
	return ((h << 8) | Hand64Canonical(c)).Exemplar(n + 1), true
}

func (h Hand64) String(n int) string {
	var s []string
	for i := 0; i < n; i++ {
		s = append(s, h.Card(i).String())
	}
	return strings.Join(s, " ")
}

func (h Hand64) Card(i int) Card {
	return Card(h >> (8 * i) & 0xff)
}

type canonSuit struct {
	s     Suit   // the original suit
	cards uint16 // bitmap of ranks
	n     int
}

// Exemplar returns one example hand of n cards that
// canonicalizes to h.
func (hc Hand64Canonical) Exemplar(n int) Hand64 {
	var suits uint
	h := Hand64(hc)
	for i := 0; i < n; i++ {
		s := h.Card(i).Suit()
		if s != XSuit {
			suits |= 1 << s
		}
	}
	ns := 0
	for i := 0; i < n; i++ {
		if h.Card(i).Suit() != XSuit {
			continue
		}
		for (suits>>ns)&1 == 1 {
			ns = (ns + 1) & 3
		}
		r := int((h >> (8 * i))) &^ (3 + 128)
		h &^= Hand64(0xff << (8 * i))    // clear i'th card
		h |= Hand64((r | ns) << (8 * i)) // set new card, with specific suit.
		ns = (ns + 1) & 3                // use a different suit next time.
	}
	return h
}

// SuitTransform represents a mapping of suits to other suits.
type SuitTransform [4]uint8

// Compose generates a suit transform that performs one suit transform after another.
// st.Compose(st2) applied to a suit s is the same as applying st first,
// and then st2 to the result.
func (st SuitTransform) Compose(st2 SuitTransform) SuitTransform {
	return SuitTransform{st2[st[0]], st2[st[1]], st2[st[2]], st2[st[3]]}
}

func (st SuitTransform) Apply(c Card) Card {
	return Card(st[c&3]) | (c &^ 3)
}

// Canonical takes an n-card Hand64, and returns its
// canonical form.
func (h Hand64) Canonical(n int) Hand64Canonical {
	r, _ := h.CanonicalWithTransform(n)
	return r
}

func (h Hand64) CanonicalWithTransform(n int) (Hand64Canonical, SuitTransform) {
	var csuits [4]canonSuit
	for i := 0; i < 4; i++ {
		csuits[i].s = Suit(i)
	}
	for i := 0; i < n; i++ {
		ci := h.Card(i)
		si := ci.Suit()
		ri := ci.RawRank()
		csuits[si].cards |= 1 << ri
		csuits[si].n++
	}
	// sort by number of cards, then by the specific
	// cards in the suit.
	sort.Slice(csuits[:], func(i, j int) bool {
		if csuits[i].n != csuits[j].n {
			return csuits[i].n > csuits[j].n
		}
		return csuits[i].cards > csuits[j].cards
	})

	var hs Hand64Canonical
	var si [4]int
	nextSuit := 0
	for i := 0; i < 4; i++ {
		if csuits[i].n+(7-n) < 5 {
			si[i] = int(XSuit)
		} else {
			si[i] = nextSuit
			nextSuit++
		}
	}

	for jj := 0; jj < 13; jj++ {
		for i := 0; i < 4; i++ {
			if (csuits[i].cards>>jj)&1 == 0 {
				continue
			}
			cr := (jj + 1) % 13
			card := (cr << 2) | si[i]
			if si[i] == int(XSuit) {
				card = (cr << 2) | 128
			}
			hs = (hs << 8) | Hand64Canonical(card)
		}
	}
	xf := SuitTransform{}
	for i := 0; i < 4; i++ {
		if si[i] > 3 {
			// We remap suits that will be x-suits
			// into spades. There can never be more than
			// 3 flushing suits (in a rainbow 3-card hand).
			// With 4 cards there can be 2 flushing suits,
			// and with 5 cards only 1 flushing suit.
			xf[int(csuits[i].s)] = 3
		} else {
			xf[int(csuits[i].s)] = uint8(si[i])
		}
	}
	return hs, xf
}
