package poker

import (
	"log"
	"sort"
	"strings"
)

// hand64 is a hand with up to N<=7 cards in it, stored
// in the lowest N bytes.
type hand64 uint64

// hand64Canonical is a compressed version of hand64.
// It is like a hand64 except that the cards are sorted, and the suits are
// canonicalized to appear in increasing order, so that the first card
// is always a club, the second a club
// or diamond, the third a club, diamond, or heart (only if
// the second card was a diamond), and so on.
// If there's no way to extend the hand to 7 cards so that
// a particular suit can form a flush, then that suit will
// be coalesced with all other suits that can't form flushes.
type hand64Canonical uint64

// CardsN returns the first n cards in the hand.
// If n is more than 7, only the first 7 cards are returned.
func (h hand64) CardsN(n int) []Card {
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
func (h hand64Canonical) Add(n int, c Card) (hand64, bool) {
	rc := 0
	for i := 0; i < n; i++ {
		ci := hand64(h).Card(i)
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
	return ((h << 8) | hand64Canonical(c)).exemplar(n+1, true), true
}

func (h hand64) String(n int) string {
	var s []string
	for i := 0; i < n; i++ {
		s = append(s, h.Card(i).String())
	}
	return strings.Join(s, " ")
}

func (h hand64) Card(i int) Card {
	return Card(h >> (8 * i) & 0xff)
}

type canonSuit struct {
	s     Suit   // the original suit
	cards uint16 // bitmap of ranks
	n     int
}

func (h hand64) SwapCards(i, j int) hand64 {
	c0 := h.Card(i)
	c1 := h.Card(j)
	h &^= 0xff << (8 * i)
	h &^= 0xff << (8 * j)
	h |= hand64(c0) << (8 * j)
	h |= hand64(c1) << (8 * i)
	return h
}

func (hc hand64Canonical) Sorted(n int) hand64Canonical {
	h := hand64(hc)
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if h.Card(i) < h.Card(j) {
				h = h.SwapCards(i, j)
			}
		}
	}
	return hand64Canonical(h)
}

func (hc hand64Canonical) exemplar(n int, botNew bool) hand64 {
	var suits uint
	h := hand64(hc)
	for i := 0; i < n; i++ {
		s := h.Card(i).Suit()
		if s != xSuit && (i > 0 || !botNew) {
			suits |= 1 << s
		}
	}
	ns := 0
	botCard := h.Card(0)
	for i := 0; i < n; i++ {
		if h.Card(i).Suit() != xSuit {
			continue
		}
		var nc Card = 0xff
		r := int((h >> (8 * i))) &^ (3 + 128)
		for tries := 0; tries < 4; tries++ {
			for (suits>>ns)&1 == 1 {
				ns = (ns + 1) & 3
			}
			xnc := Card(r | ns)
			if !botNew || xnc != botCard {
				nc = xnc
				break
			}
			ns = (ns + 1) & 3
		}
		if nc == 0xff {
			log.Printf("suits = %x", suits)
			log.Fatalf("exemplar(%s, %d, %v) failed to find suit at step %d", hand64(hc).String(n), n, botNew, i)
		}
		h &^= hand64(0xff << (8 * i)) // clear i'th card
		h |= hand64(nc) << (8 * i)    // set new card, with specific suit.
		ns = (ns + 1) & 3             // use a different suit next time.
	}
	return h
}

// Exemplar returns one example hand of n cards that
// canonicalizes to h.
func (hc hand64Canonical) Exemplar(n int) hand64 {
	return hc.exemplar(n, false)
}

// suitTransform represents a mapping of suits to other suits.
type suitTransform [4]uint8
type suitTransformByte uint8

// Compose generates a suit transform that performs one suit transform after another.
// st.Compose(st2) applied to a suit s is the same as applying st first,
// and then st2 to the result.
func (st suitTransform) Compose(st2 suitTransform) suitTransform {
	return suitTransform{st2[st[0]], st2[st[1]], st2[st[2]], st2[st[3]]}
}

func (st suitTransform) Byte() suitTransformByte {
	if st[0] > 3 || st[1] > 3 || st[2] > 3 || st[3] > 3 {
		log.Fatalf("can't transform suit transform %v to byte", st[:])
	}
	return suitTransformByte(st[0] | (st[1] << 2) | (st[2] << 4) | (st[3] << 6))
}

func (st suitTransformByte) apply(c Card) Card {
	return Card(st>>(2*(c&3))&3) | (c &^ 3)
}

var applyTable = func() [256 * 64]Card {
	var r [256 * 64]Card
	for i := 0; i < 256; i++ {
		for j := 0; j < 52; j++ {
			r[64*i+j] = suitTransformByte(i).apply(Card(j))
		}
	}
	return r
}()

func (st suitTransformByte) Apply(c Card) Card {
	return applyTable[int(st)*64+int(c)]
}

func (st suitTransformByte) Compose(st2 suitTransformByte) suitTransformByte {
	return composeTable[int(st)*256+int(st2)]
}

func (st suitTransformByte) compose(st2 suitTransformByte) suitTransformByte {
	var r suitTransformByte
	r = (st2 >> (2 * (st & 3))) & 3
	r |= ((st2 >> (2 * ((st >> 2) & 3))) & 3) << 2
	r |= ((st2 >> (2 * ((st >> 4) & 3))) & 3) << 4
	r |= ((st2 >> (2 * ((st >> 6) & 3))) & 3) << 6
	return r
}

var composeTable = func() [256 * 256]suitTransformByte {
	var r [256 * 256]suitTransformByte
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			r[i*256+j] = suitTransformByte(i).compose(suitTransformByte(j))
		}
	}
	return r
}()

func (st suitTransformByte) Long() suitTransform {
	return suitTransform{uint8(st & 3), uint8((st >> 2) & 3), uint8((st >> 4) & 3), uint8((st >> 6) & 3)}
}

var suitTransformByteIdentity = suitTransform{0, 1, 2, 3}.Byte()

func (st suitTransform) Apply(c Card) Card {
	return Card(st[c&3]) | (c &^ 3)
}

// Canonical takes an n-card hand64, and returns its
// canonical form.
func (h hand64) Canonical(n, finalN int) hand64Canonical {
	r, _ := h.CanonicalWithTransform(n, finalN)
	return r
}

func (h hand64) CanonicalWithTransform(n, finalN int) (hand64Canonical, suitTransform) {
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

	var hs hand64Canonical
	var si [4]int
	nextSuit := 0
	for i := 0; i < 4; i++ {
		if csuits[i].n+(finalN-n) < 5 {
			si[i] = int(xSuit)
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
			if si[i] == int(xSuit) {
				card = (cr << 2) | 128
			}
			hs = (hs << 8) | hand64Canonical(card)
		}
	}
	xf := suitTransform{}
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
