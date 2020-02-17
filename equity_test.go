package poker

import (
	"fmt"
	"math"
	"testing"
)

type eqTest struct {
	name  string
	hands [][2]Card
	board []Card

	wantBoards int
	wantEqs    []Equity
}

func parseHand2(s string) ([2]Card, error) {
	var hz [2]Card
	if len(s) != 4 {
		return hz, fmt.Errorf("expect hand in format like CAKH, got %q", s)
	}
	c0, ok0 := NameToCard[s[:2]]
	c1, ok1 := NameToCard[s[2:]]
	if !ok0 || !ok1 {
		return hz, fmt.Errorf("failed to parse hand %q", s)
	}
	return [2]Card{c0, c1}, nil
}

func TestEquity(t *testing.T) {
	hand := func(s string) [2]Card {
		h, err := parseHand2(s)
		if err != nil {
			t.Fatal(err)
		}
		return h
	}
	card := func(s string) Card {
		c, ok := NameToCard[s]
		if !ok {
			t.Fatalf("can't parse card %s", s)
		}
		return c
	}
	tcs := []eqTest{
		eqTest{
			name:       "preflop AcKh vs KdTh",
			hands:      [][2]Card{hand("CAHK"), hand("DKHT")},
			wantBoards: 48 * 47 * 46 * 45 * 44 / (5 * 4 * 3 * 2),
			wantEqs: []Equity{
				{Win: 0.7366, Tie: 0.0114, Equity: 0.7366 + 0.5*0.0114},
				{Win: 0.252, Tie: 0.0114, Equity: 0.252 + 0.5*0.0114},
			},
		},
		eqTest{
			name:       "AcKh vs KdTh vs 9h9d on flop 2d2h2s",
			hands:      [][2]Card{hand("CAHK"), hand("DKHT"), hand("H9D9")},
			board:      []Card{card("D2"), card("H2"), card("S2")},
			wantBoards: (52 - 6 - 3) * (52 - 6 - 3 - 1) / 2,
			wantEqs: []Equity{
				{Win: 0.1694, Tie: 0.0952, Equity: 0.1694 + 0.33*0.0166 + 0.5*(0.0952-0.0166)},
				{Win: 0.1096, Tie: 0.0952, Equity: 0.1096 + 0.33*0.0166 + 0.5*(0.0952-0.0166)},
				{Win: 0.6257, Tie: 0.0166, Equity: 0.6257 + 0.33*0.0166},
			},
		},
		eqTest{
			name:       "AcKh vs KdTh vs 9h9d on river 2d2h2s Ks Jc",
			hands:      [][2]Card{hand("CAHK"), hand("DKHT"), hand("H9D9")},
			board:      []Card{card("D2"), card("H2"), card("S2"), card("SK"), card("CJ")},
			wantBoards: 1,
			wantEqs: []Equity{
				{Tie: 1.0, Equity: 0.5},
				{Tie: 1.0, Equity: 0.5},
				{},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			eqs, err := HoldemEquities(tc.hands, tc.board)
			if err != nil {
				t.Fatalf("failed to compute equities: %v", err)
			}
			for i := range tc.hands {
				got := eqs[i]
				gotEq := got.Equity
				gotWin := got.Win
				gotTie := got.Tie
				gotBoards := got.Boards
				want := tc.wantEqs[i]
				if gotBoards != tc.wantBoards {
					t.Errorf("hand %s: expected %d boards, got %d", tc.hands[i], tc.wantBoards, gotBoards)
				}
				if math.Abs(gotEq-want.Equity) > 0.001 {
					t.Errorf("hand %s: equity=%f, want %f", tc.hands[i], gotEq, want.Equity)
				}
				if math.Abs(gotWin-want.Win) > 0.001 {
					t.Errorf("hand %s: win=%f, want %f", tc.hands[i], gotWin, want.Win)
				}
				if math.Abs(gotTie-want.Tie) > 0.001 {
					t.Errorf("hand %s: tie=%f, want %f", tc.hands[i], gotTie, want.Tie)
				}
			}
		})
	}

}
