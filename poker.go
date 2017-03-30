// Package poker provides functions for poker tools.
package poker

import (
	"fmt"
	"log"
)

// A Card is a single playing card.
// The top two bits are the suit, and the lowest 6 bits
// store the (r-1)th prime number.
// This representation enables fast hand evaluation.
// Heavily based on the ideas from Cactus Pete's poker
// hand evaluator, which can be found here:
// See http://www.suffecool.net/poker/evaluator.html
type Card uint16

// Suit returns the suit of a card.
func (c Card) Suit() Suit {
	switch c >> 6 {
	case 1:
		return Club
	case 2:
		return Diamond
	case 4:
		return Heart
	case 8:
		return Spade
	}
	log.Fatalf("unknown suit: %d", c>>6)
	return Club
}

// Rank returns the rank of a card.
func (c Card) Rank() Rank {
	r := uint16(c & 0x3f)
	for pr := Rank(1); pr <= 13; pr++ {
		if primes[pr-1] == r {
			return pr
		}
	}
	return 0
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
)

var suits = map[Suit]string{
	Club:    "C",
	Diamond: "D",
	Heart:   "H",
	Spade:   "S",
}

func (s Suit) String() string {
	return suits[s]
}

// A Rank describes the rank of a card: A23456789TJQK.
// Ace is 1, King is 13.
type Rank int

func (r Rank) String() string {
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
	return Card(uint16(1<<(6+s)) | primes[r-1]), nil
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
