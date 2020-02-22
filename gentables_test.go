package poker

import (
	"testing"
)

type tableTestCase struct {
	hand string
}

func TestTables(t *testing.T) {
	tcs := []tableTestCase{
		{hand: "HK DK S2 D3 CQ DJ D7"},
		{hand: "SA HA DA DK HK SQ CA"},
		{hand: "SA SQ ST DT S5 S3 CA"},
		{hand: "SA SK SQ SJ ST S9 S8"},
		{hand: "SA SK SQ CJ ST S9 S8"},
	}
	for _, tc := range tcs {
		t.Run(tc.hand, func(t *testing.T) {
			h, err := parseHand(tc.hand)
			if err != nil {
				t.Fatal(err)
			}
			var cards [7]Card
			copy(cards[:], h)
			gotRank := NodeEval7(&cards)
			wantRank := Eval7(&cards)
			if gotRank != wantRank {
				t.Errorf("NodeEval7() = %d, want %d", gotRank, wantRank)
			}
		})
	}
}
