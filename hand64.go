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
	cards uint16 // bitmap of ranks
	n     int
}

// Examplar returns one example hand of n cards that
// canonicalizes to h.
func (hc Hand64Canonical) Examplar(n int) Hand64 {
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

// Canonical takes an n-card Hand64, and returns its
// canonical form.
func (h Hand64) Canonical(n int) Hand64Canonical {
	var csuits [4]canonSuit
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
	return hs
}
