package poker

import (
	"math/rand"
	"testing"
)

type tableTestCase struct {
	hand string
}

func TestEval5Single(t *testing.T) {
	cases := []string{
		"S2 H2 D2 C2 CA",
		"CT DT HT ST D8",
	}
	var cards [5]Card
	for _, c := range cases {
		h, err := parseHand(c)
		if err != nil {
			t.Fatalf("parse error of %s: %v", c, err)
		}
		if len(h) != 5 {
			t.Fatalf("got %d cards when parsing %s: want 5", len(h), c)
		}
		for perms := 0; perms < 120; perms++ {
			copy(cards[:], h)
			p := perms
			for i := 0; i < 5; i++ {
				cards[i], cards[i+perms%(5-i)] = cards[i+perms%(5-i)], cards[i]
				p /= (5 - i)
			}
			gotEval := NodeEval5(&cards)
			wantEval := EvalSlow(cards[:])
			if gotEval != wantEval {
				t.Errorf("%v.NodeEval() = %d, want %d", cards[:], gotEval, wantEval)
				t.Errorf("... hand evaluated to %v", evalInfo.rankTo5[gotEval])
			}
		}
	}
}

func TestEval5(t *testing.T) {
	fails := 0
	skipped := 0
	const failThreshold = 20
	const failSample = 100
	for a := Card(0); a < Card(52); a++ {
		for b := Card(a) + 1; b < Card(52); b++ {
			for c := Card(b) + 1; c < Card(52); c++ {
				for d := Card(c) + 1; d < Card(52); d++ {
					for e := Card(d) + 1; e < Card(52); e++ {
						wantEval := EvalSlow([]Card{a, b, c, d, e})
						for perms := 0; perms < 120; perms += 10 {
							p := perms
							h := [5]Card{a, b, c, d, e}

							for i := 0; i < 5; i++ {
								h[i], h[i+perms%(5-i)] = h[i+perms%(5-i)], h[i]
								p /= (5 - i)
							}
							gotEval := Eval5(&h)
							if gotEval != wantEval {
								if fails >= failThreshold && rand.Intn(failSample) != 0 {
									fails++
									skipped++
									continue
								}
								if skipped > 0 {
									t.Errorf("[skipped reporting %d failures]", skipped)
									skipped = 0
								}
								t.Errorf("%v.Eval() = %d, want %d", h[:], gotEval, wantEval)
								t.Errorf("... hand evaluated to %v", evalInfo.rankTo5[gotEval])
								fails++
								if fails == failThreshold {
									t.Errorf("too many failures: now reporting 1/%d", failSample)
								}
							}
						}
					}
				}
			}
		}
	}
	if skipped > 0 {
		t.Errorf("[skipped reporting %d failures]", skipped)
	}
}

func BenchmarkEval5(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var T int64
		for a := Card(0); a < Card(52); a++ {
			for b := Card(a) + 1; b < Card(52); b++ {
				for c := Card(b) + 1; c < Card(52); c++ {
					for d := Card(c) + 1; d < Card(52); d++ {
						for e := Card(d) + 1; e < Card(52); e++ {
							h := [5]Card{a, b, c, d, e}
							T += int64(Eval5(&h))
						}
					}
				}
			}
		}
		// make sure we're not optimizing the code away.
		if T == 0 {
			panic("x")
		}
	}
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
		h, err := parseHand(tc.hand)
		if err != nil {
			t.Fatalf("%s: parseHand failed: %v", tc.hand, err)
		}
		var cards [7]Card
		copy(cards[:], h)
		gotRank := NodeEval7(&cards)
		wantRank := EvalSlow7(&cards)
		if gotRank != wantRank {
			t.Errorf("%s: NodeEval7() = %d, want %d", tc.hand, gotRank, wantRank)
		}
	}
}
