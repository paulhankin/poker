// Package holdem contains code for analysing holdem
// hands and plays.
package holdem

import (
	"github.com/paulhankin/poker"
)

// A Hand is a single player's hole cards.
// The rank of the first card should be larger or equal to
// the rank of the second. If equal, the suit of the first
// card should be higher than the second.
type Hand [2]poker.Card

// A Range is a distribution of hole cards
// that a player is modelled to have at a given
// point in the play.
type Range interface {
	// P is the probability that the player has this particular
	// hand. The return value must be between 0 and 1 inclusive.
	P(h Hand) float64
}

// A SimpleRange describes a range that distinguishes between
// unsuited and suited hands, but not between specific suits.
// Paired hands (with RawRank i) are stored at i, i.
// Suited hands (with RawRank's i>j) are stored at i, j (i>j)
// Offsuit hands (with RawRank's i>j) are stored at j, i
type SimpleRange [13][13]float64

// P returns the probability that the given hand is in the
// SimpleRange.
func (s *SimpleRange) P(h Hand) float64 {
	r0 := h[0].RawRank()
	r1 := h[1].RawRank()
	if h[0].Suit() != h[1].Suit() {
		r0, r1 = r1, r0
	}
	return s[r0][r1]
}

// A MapRange is a map from hands to probabilities.
type MapRange map[Hand]float64

// P returns the probability that a given hand is in the MapRange.
func (mr MapRange) P(h Hand) float64 {
	return mr[h]
}
