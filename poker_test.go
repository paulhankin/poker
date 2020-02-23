package poker

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func parseHand(s string) ([]Card, error) {
	cs := strings.Split(s, " ")
	r := make([]Card, len(cs))
	for i, p := range cs {
		var ok bool
		r[i], ok = NameToCard[p]
		if !ok {
			return nil, fmt.Errorf("can't parse card %s", p)
		}
	}
	return r, nil
}

func TestDescriptions(t *testing.T) {
	// Hands and their long and short descriptions.
	// When the short description is expected to be the same as the long,
	// it's omitted.
	hands := []struct {
		hand                string
		wantLong, wantShort string
	}{
		{hand: "HA HK HQ HJ HT", wantLong: "A straight flush"},
		{hand: "D5 D4 D3 D2 DA", wantLong: "5 straight flush"},
		{hand: "HA SA DA CA C3", wantLong: "AAAA-3", wantShort: "AAAA-x"},
		{hand: "HT ST DT CT D2", wantLong: "TTTT-2", wantShort: "TTTT-x"},
		{hand: "H2 S2 D2 C2 CK", wantLong: "2222-K", wantShort: "2222-x"},
		{hand: "SK HK DK C2 H2", wantLong: "KKK-22", wantShort: "KKK-xx"},
		{hand: "ST HT CT CA DA", wantLong: "TTT-AA", wantShort: "TTT-xx"},
		{hand: "HA HK HQ H2 H3", wantLong: "AKQ32 flush"},
		{hand: "HA HQ H8 H7 H5", wantLong: "AQ875 flush"},
		{hand: "SK SJ S9 S7 S5", wantLong: "KJ975 flush"},
		{hand: "CA SK SQ SJ ST", wantLong: "A straight"},
		{hand: "HK HQ DJ CT S9", wantLong: "K straight"},
		{hand: "H6 D5 C4 D3 H2", wantLong: "6 straight"},
		{hand: "H5 D4 C3 D2 CA", wantLong: "5 straight"},
		{hand: "HA DA CA C3 D2", wantLong: "AAA-3-2", wantShort: "AAA-x-y"},
		{hand: "HQ DQ CQ C3 D2", wantLong: "QQQ-3-2", wantShort: "QQQ-x-y"},
		{hand: "H2 D2 C2 CA DK", wantLong: "222-A-K", wantShort: "222-x-y"},
		{hand: "HA DA CK DK H3", wantLong: "AA-KK-3"},
		{hand: "HA DA CQ DQ H4", wantLong: "AA-QQ-4"},
		{hand: "HT DT C8 D8 D2", wantLong: "TT-88-2"},
		{hand: "H9 D9 C7 D7 CA", wantLong: "99-77-A"},
		{hand: "HA DA CK DQ D2", wantLong: "AA-K-Q-2"},
		{hand: "HA DA CQ DJ D7", wantLong: "AA-Q-J-7"},
		{hand: "HK DK CQ DJ D7", wantLong: "KK-Q-J-7"},
		{hand: "H2 D2 CA DK HQ", wantLong: "22-A-K-Q"},
		{hand: "SA HQ H8 H7 H5", wantLong: "A-Q-8-7-5"},
		{hand: "DK SJ S9 S7 S5", wantLong: "K-J-9-7-5"},
		{hand: "S7 D5 H4 S3 S2", wantLong: "7-5-4-3-2"},
		{hand: "HA SA DA", wantLong: "AAA"},
		{hand: "S5 C5 D5", wantLong: "555"},
		{hand: "DA CA D3", wantLong: "AA-3"},
		{hand: "DT CT HK", wantLong: "TT-K"},
		{hand: "HA HQ H2", wantLong: "A-Q-2"},
		{hand: "H5 H2 H3", wantLong: "5-3-2"},
		{hand: "HK DK S2 D3 CQ DJ D7", wantLong: "KK-Q-J-7"},
		{hand: "SA HA DA DK HK SQ CA", wantLong: "AAAA-K", wantShort: "AAAA-x"},
		{hand: "SA SQ ST DT S5 S3 CA", wantLong: "AQT53 flush"},
	}
	for i := range hands {
		h0, err := parseHand(hands[i].hand)
		if err != nil {
			t.Fatalf("parseHand(%s) gave error %s", hands[i].hand, err)
		}
		for perms := 0; perms < 10; perms++ {
			// Randomly permute the hands.
			for i := 0; i < len(h0)-1; i++ {
				j := i + rand.Intn(len(h0)-1-i)
				h0[i], h0[j] = h0[j], h0[i]
			}
			ld, err := Describe(h0)
			if err != nil {
				t.Fatalf("Describe(%s) produced an error: %s", Hand(h0), err)
			}
			if ld != hands[i].wantLong {
				t.Errorf("Describe(%s) = %s, want %s", Hand(h0), ld, hands[i].wantLong)
			}
			wantShort := hands[i].wantShort
			if wantShort == "" {
				wantShort = hands[i].wantLong
			}
			sd, err := DescribeShort(h0)
			if err != nil {
				t.Fatalf("DescribeShort(%s) produced an error: %s", Hand(h0), err)
			}
			if sd != wantShort {
				t.Errorf("DescribeShort(%s) = %s, want %s", Hand(h0), sd, wantShort)
			}
		}
	}
}

func TestRankings(t *testing.T) {
	// These hands are in descending order of strength.
	// 3-card hands should in general rank lower than 5-card
	// hands.
	hands := []string{
		"HA HK HQ HJ HT",
		"D5 D4 D3 D2 DA",
		"HA SA DA CA C3",
		"HT ST DT CT D2",
		"H2 S2 D2 C2 CK",
		"SK HK DK C2 H2",
		"ST HT CT CA DA",
		"HA HK HQ H2 H3",
		"HA HQ H8 H7 H5",
		"SK SJ S9 S7 S5",
		"CA SK SQ SJ ST",
		"HK HQ DJ CT S9",
		"H6 D5 C4 D3 H2",
		"H5 D4 C3 D2 CA",
		"HA DA CA C3 D2",
		"HA DA CA",
		"HQ DQ CQ C3 D2",
		"HQ DQ CQ",
		"HJ DJ CJ",
		"H2 D2 C2 CA DK",
		"H2 D2 C2",
		"HA DA CK DK H3",
		"HA DA CQ DQ H4",
		"HT DT C8 D8 D2",
		"H9 D9 C7 D7 CA",
		"HA DA CK DQ D2",
		"HA DA CK",
		"HA DA CQ DJ D7",
		"HA DA CQ",
		"HK DK CQ DJ D7",
		"HK DK CQ",
		"H2 D2 CA DK HQ",
		"H2 D2 CA",
		"SA HQ H9",
		"SA HQ H8 H7 H5",
		"SA HQ H8",
		"DK SJ S9 S7 S5",
		"DK SJ S9",
		"S7 D5 H4 S3 S2",
		"S7 D5 H4",
	}
	prevEV := int16(0x7fff)
	prevHand := ""
	for i := range hands {
		h := hands[i]
		h0, err := parseHand(h)
		if err != nil {
			t.Fatalf("parseHand(%s) gave error %s", hands[i], err)
		}
		ev := Eval(h0)
		if ev >= prevEV {
			t.Errorf("Expected %s to beat %s, but got scores %d and %d", prevHand, hands[i], prevEV, ev)
		}
		prevEV, prevHand = ev, hands[i]
	}
}

func TestToHand(t *testing.T) {
	hands := []string{
		"HA HK HQ HJ HT",
		"D5 D4 D3 D2 DA",
		"HA SA DA CA C3",
		"HT ST DT CT D2",
		"H2 S2 D2 C2 CK",
		"SK HK DK C2 H2",
		"ST HT CT CA DA",
		"HA HK HQ H2 H3",
		"HA HQ H8 H7 H5",
		"SK SJ S9 S7 S5",
		"CA SK SQ SJ ST",
		"HK HQ DJ CT S9",
		"H6 D5 C4 D3 H2",
		"H5 D4 C3 D2 CA",
		"HA DA CA C3 D2",
		"HA DA CA",
		"HQ DQ CQ C3 D2",
		"HQ DQ CQ",
		"HJ DJ CJ",
		"H2 D2 C2 CA DK",
		"H2 D2 C2",
		"HA DA CK DK H3",
		"HA DA CQ DQ H4",
		"HT DT C8 D8 D2",
		"H9 D9 C7 D7 CA",
		"HA DA CK DQ D2",
		"HA DA CK",
		"HA DA CQ DJ D7",
		"HA DA CQ",
		"HK DK CQ DJ D7",
		"HK DK CQ",
		"H2 D2 CA DK HQ",
		"H2 D2 CA",
		"SA HQ H9",
		"SA HQ H8 H7 H5",
		"SA HQ H8",
		"DK SJ S9 S7 S5",
		"DK SJ S9",
		"S7 D5 H4 S3 S2",
		"S7 D5 H4",
	}
	for i := range hands {
		h := hands[i]
		h0, err := parseHand(h)
		if err != nil {
			t.Fatalf("parseHand(%s) gave error %s", hands[i], err)
		}
		h1 := []Card{}
		ok := false
		if len(h0) == 3 {
			h1, ok = EvalToHand3(Eval(h0))
		} else {
			h1, ok = EvalToHand5(Eval(h0))
		}
		if !ok {
			t.Fatalf("EvalToHand(%s) failed. Eval=%d", h0, Eval(h0))
		}
		want, err := DescribeShort(h0)
		if err != nil {
			t.Fatalf("DescribeShort(%s) gave error %s", h0, err)
		}
		got, err := DescribeShort(h1)
		if err != nil {
			t.Fatalf("DescribeShort(%s) gave error %s", h1, err)
		}
		if got != want {
			t.Errorf("EvalToHand(%s) = %s, want %s", h0, got, want)
		}
	}
}
