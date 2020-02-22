// Package poker provides functions for poker tools.
package poker

import (
	"fmt"
	"log"
)

// A Card is a single playing card. It's represented as a
// number from 0 to 51. The bottom two bits are the suit.
//
// In some cases, a special "x" suit is used, which represents
// some anonymized suit. This x is used in hand64 when canonicalizing
// hands. The choice of "x" in a hand has constraints: each
// card must be a suit different from the named suits in a hand,
// two cards of the same rank with x-suit must have different
// suits, and there can't be too many cards of the same suit, so that
// if you add 7-N more cards of that suit, then there can't be a flush
// in that suit.
// Eg: in KhQhTx8x, the T, 8 must be different non-heart suits.
// In AxKxQxJx, all four cards must be different suits.
type Card uint8

// Suit returns the suit of a card.
func (c Card) Suit() Suit {
	if c > 127 {
		return XSuit
	}
	return Suit(c & 3)
}

// Rank returns the rank of a card. It returns 0
// if the card isn't valid.
func (c Card) Rank() Rank {
	r := (0xf & Rank(c>>2)) + 1
	if r > 13 {
		return 0
	}
	return r
}

// Valid reports whether the card is a valid card.
func (c Card) Valid() bool {
	return c < 52
}

// RawRank returns a number from 0 to 12 representing the
// strength of the card. 2->0, 3->1, ..., K->11, A->12.
func (c Card) RawRank() int {
	return (int(c.Rank()) + 11) % 13
}

func (c Card) String() string {
	return c.Suit().String() + c.Rank().String()
}

// A Suit is a suit: clubs, diamonds, hearts or spades.
type Suit uint8

const (
	Club    = Suit(0)
	Diamond = Suit(1)
	Heart   = Suit(2)
	Spade   = Suit(3)

	XSuit   = Suit(254)
	BadSuit = Suit(255)
)

var suits = map[Suit]string{
	Club:    "C",
	Diamond: "D",
	Heart:   "H",
	Spade:   "S",
	XSuit:   "x",
}

func (s Suit) String() string {
	r, ok := suits[s]
	if !ok {
		return "?"
	}
	return r
}

// A Rank describes the rank of a card: A23456789TJQK.
// Ace is 1, King is 13.
type Rank int

func (r Rank) String() string {
	if r < 1 || r > 13 {
		return "?"
	}
	return "A23456789TJQK"[r-1 : r]
}

var primes = []uint16{
	2, 3, 5, 7, 11, 13, 17, 23, 29, 31, 37, 41, 43,
}

// MakeCard constructs a card from a suit and rank.
func MakeCard(s Suit, r Rank) (Card, error) {
	if s > 3 || r == 0 || r > 13 {
		return 0, fmt.Errorf("illegal card %d %d", s, r)
	}
	return Card(r-1)*4 + Card(s), nil
}

// NameToCard maps card names (for example, "C8" or "HA") to a card value.
var NameToCard = map[string]Card{}

// Cards is a full deck of all cards. Sorted by suit and then rank.
var Cards []Card

func init() {
	for s := Suit(0); s <= Suit(3); s++ {
		for r := Rank(1); r <= Rank(13); r++ {
			c, err := MakeCard(s, r)
			if err != nil {
				log.Fatalf("Cards construction failed: %s", err)
			}
			NameToCard[c.String()] = c
			Cards = append(Cards, c)
		}
	}
}
