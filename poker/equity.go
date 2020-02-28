package poker

import (
	"fmt"
	"sort"
	"strings"
)

// Equity contains information about poker hand equity.
type Equity struct {
	Equity float64 // total equity in the pot
	Win    float64 // equity gained from outright winning the pot
	Tie    float64 // probability of tieing with 1 or more hands
	Boards int     // how many runouts were computed
}

func boardString(b []Card) string {
	var parts []string
	for _, c := range b {
		parts = append(parts, c.String())
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func holdemRiverEquities(hbs [][7]Card, evs []int16, eqs []Equity) {
	H := len(hbs)
	winCount := 0
	var bestEV int16 = -1000
	for i := 0; i < H; i++ {
		ev := Eval7(&hbs[i])
		evs[i] = ev
		if ev > bestEV {
			winCount = 1
			bestEV = ev
		} else if ev == bestEV {
			winCount++
		}
	}

	v := 1.0 / float64(winCount)
	for i := 0; i < H; i++ {
		if evs[i] != bestEV {
			continue
		}
		eqs[i].Equity += v
		if winCount == 1 {
			eqs[i].Win += v
		} else {
			eqs[i].Tie += 1.0
		}
	}
}

func getRemainingDeck(hands [][2]Card, board []Card) ([]Card, error) {
	got := map[Card]int{}
	for i, h := range hands {
		if !h[0].Valid() {
			return nil, fmt.Errorf("hand %d contains invalid first card %d", i, h[0])
		}
		if !h[1].Valid() {
			return nil, fmt.Errorf("hand %d contains invalid second card %d", i, h[1])
		}
		got[h[0]]++
		got[h[1]]++
	}
	for i, b := range board {
		if !b.Valid() {
			return nil, fmt.Errorf("board[%d] card is invalid: %d", i, b)
		}
		got[b]++
	}
	if len(got) != 2*len(hands)+len(board) {
		var dups []string
		for c, i := range got {
			if i > 1 {
				dups = append(dups, c.String())
			}
		}
		sort.Strings(dups)
		return nil, fmt.Errorf("duplicate cards: %v found", dups)
	}
	if len(board) > 5 {
		return nil, fmt.Errorf("board %s has more than 5 (%d) cards", boardString(board), len(board))
	}

	// deck is all the cards that aren't already in a hand or board.
	var deck []Card
	for _, c := range Cards {
		if got[c] > 0 {
			continue
		}
		deck = append(deck, c)
	}
	return deck, nil
}

// HoldemEquities returns the river equities for the given holdem hands
// given a board of up to 5 cards.
// The hands and board must be distinct, and the board can't have more
// than 5 cards in it.
func HoldemEquities(hands [][2]Card, board []Card) ([]Equity, error) {
	deck, err := getRemainingDeck(hands, board)
	if err != nil {
		return nil, err
	}

	// hbs is hands and board.
	// We store the fixed cards (given hand and board) at the
	// end of the array for convenience when we're modifying
	// values in the eval loop.
	hbs := make([][7]Card, len(hands))
	for i, h := range hands {
		hbs[i][7-len(board)-2+0] = h[0]
		hbs[i][7-len(board)-2+1] = h[1]
		for j, b := range board {
			hbs[i][7-len(board)-2+2+j] = b
		}
	}

	eqs := make([]Equity, len(hands))
	evs := make([]int16, len(hands))

	if len(board) == 5 {
		holdemRiverEquities(hbs, evs, eqs)
		for i := range eqs {
			eqs[i].Boards = 1
		}
		return eqs, nil
	}

	idxs := make([]int, 5-len(board))
	for i := range idxs {
		idxs[i] = i
	}

	H := len(hands)
	T := 0 // total number of runouts we've considered.

	for {
		T++
		// update the boards
		for j, ix := range idxs {
			c := deck[ix]
			for i := 0; i < H; i++ {
				hbs[i][j] = c
			}
		}
		holdemRiverEquities(hbs, evs, eqs)
		if !incHEIndex(idxs, len(deck)) {
			break
		}
	}
	for i := range eqs {
		eqs[i].Equity /= float64(T)
		eqs[i].Win /= float64(T)
		eqs[i].Tie /= float64(T)
		eqs[i].Boards = int(T)
	}
	return eqs, nil
}

func incHEIndex(idx []int, dl int) bool {
	K := len(idx)
	// Scan right-to-left to find an index we can increase.
	// When we find it, reset indexes to the right of us.
	// Most of the time the first iteration of the loop will
	// finish.
	for i := K - 1; i >= 0; i-- {
		if idx[i] == dl-(K-1-i)-1 {
			// we can't increase this index any more
			continue
		}
		idx[i]++
		for j := i + 1; j < K; j++ {
			idx[j] = idx[i] + j - i
		}
		return true
	}
	return false
}
