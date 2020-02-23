package poker

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type addTestCase struct {
	N     int
	start string
	add   string
	want  string
}

func TestHand64Add(t *testing.T) {
	tcs := []addTestCase{
		{5, "CT HT DT ST", "C8", "S8 CT DT HT ST"},
		{5, "CT HT DT ST", "S8", "S8 CT DT HT ST"},
		{5, "H2 D2 C2 CQ", "S2", "S2 CQ D2 H2 C2"},
	}
	for _, tc := range tcs {
		ch0, err := parseHand(tc.start)
		if err != nil {
			t.Fatal(err)
		}
		want, err := parseHand(tc.want)
		if err != nil {
			t.Fatal(err)
		}
		addCards, err := parseHand(tc.add)
		if err != nil {
			t.Fatal(err)
		}
		if len(addCards) != 1 {
			t.Fatalf("expected 1 card to add, got %s", tc.add)
		}
		addCard := addCards[0]
		var h64 Hand64
		for _, c := range ch0 {
			h64 = (h64 << 8) | Hand64(c)
		}
		h64c, xf := h64.CanonicalWithTransform(len(ch0), tc.N)
		goth, ok := h64c.Add(len(ch0), xf.Apply(addCard))
		if !ok {
			t.Errorf("%v.Add(%s) failed", ch0, addCard)
			continue
		}
		got := goth.CardsN(tc.N)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s.Add(%s) = %v, want %v", Hand64(h64c).String(len(ch0)), addCard, got, want)
		}
	}
}

type canonTestCase struct {
	hand string
	want string
}

func TestCanonical(t *testing.T) {
	tcs := []canonTestCase{
		{hand: "HK DK S2 D3 CQ DJ D7", want: "xK xK xQ xJ x7 x3 x2"},
		{hand: "SA HA DA DK HK SQ CA", want: "xA xA xA xA xK xK xQ"},
		{hand: "SA SQ ST DT S5 S3 CA", want: "xA CA CQ xT CT C5 C3"},
		{hand: "SA SQ ST D9 S5 S3", want: "CA CQ CT x9 C5 C3"},
		{hand: "SA SQ ST D9 D5 S3", want: "CA CQ CT x9 x5 C3"},
		{hand: "SA SQ ST D9 D5 D3", want: "xA xQ xT x9 x5 x3"},
		{hand: "HA HK HQ HJ HT", want: "CA CK CQ CJ CT"},
		{hand: "D5 D4 D3 D2 DA", want: "CA C5 C4 C3 C2"},
		{hand: "HA SA DA CA C3", want: "xA xA xA xA x3"},
		{hand: "HT ST DT CT D2", want: "xT xT xT xT x2"},
		{hand: "H2 S2 D2 C2 CK", want: "xK x2 x2 x2 x2"},
		{hand: "SK HK DK C2 H2", want: "xK xK xK x2 x2"},
		{hand: "ST HT CT CA DA", want: "xA xA xT xT xT"},
		{hand: "HA HK HQ H2 H3", want: "CA CK CQ C3 C2"},
		{hand: "HA HQ H8 H7 H5", want: "CA CQ C8 C7 C5"},
		{hand: "SK SJ S9 S7 S5", want: "CK CJ C9 C7 C5"},
		{hand: "CA SK SQ SJ ST", want: "xA CK CQ CJ CT"},
		{hand: "HK HQ DJ CT S9", want: "xK xQ xJ xT x9"},
		{hand: "H6 D5 C4 D3 H2", want: "x6 x5 x4 x3 x2"},
		{hand: "H5 D4 C3 D2 CA", want: "xA x5 x4 x3 x2"},
		{hand: "HA DA CA C3 D2", want: "xA xA xA x3 x2"},
		{hand: "HQ DQ CQ D3 D2", want: "xQ xQ CQ C3 C2"},
		{hand: "H2 D2 C2 CA DK", want: "xA xK x2 x2 x2"},
		{hand: "HA DA CK HK H3", want: "xA CA xK CK C3"},
		{hand: "HA DA CQ DQ H4", want: "xA xA xQ xQ x4"},
		{hand: "HT DT C8 D8 D2", want: "xT CT x8 C8 C2"},
		{hand: "H9 C9 C7 D7 CA", want: "CA x9 C9 x7 C7"},
		{hand: "HA DA CK DQ D2", want: "xA CA xK CQ C2"},
		{hand: "HA DA CQ DJ D7", want: "xA CA xQ CJ C7"},
		{hand: "HK DK CQ DJ D7", want: "xK CK xQ CJ C7"},
		{hand: "H2 D2 CA DK HQ", want: "xA xK xQ x2 x2"},
		{hand: "SA HQ H8 H7 H5", want: "xA CQ C8 C7 C5"},
		{hand: "DK SJ S9 S7 S5", want: "xK CJ C9 C7 C5"},
		{hand: "S7 D5 H4 S3 S2", want: "C7 x5 x4 C3 C2"},
		{hand: "DK CQ HJ ST", want: "xK xQ xJ xT"},
		{hand: "DK DQ HJ ST", want: "CK CQ xJ xT"},
		{hand: "DK DQ HJ HT", want: "CK CQ DJ DT"},
		{hand: "SK SQ HJ ST", want: "CK CQ xJ CT"},
		{hand: "SK SQ SJ ST", want: "CK CQ CJ CT"},
		{hand: "HA SA DA", want: "HA DA CA"},
		{hand: "S5 C5 D5", want: "H5 D5 C5"},
		{hand: "DA CA D3", want: "DA CA C3"},
		{hand: "DT CT HK", want: "CK HT DT"},
		{hand: "HA HQ H2", want: "CA CQ C2"},
		{hand: "HA HQ C2", want: "CA CQ D2"},
		{hand: "H5 H2 H3", want: "C5 C3 C2"},
	}
	for _, tc := range tcs {
		t.Run(fmt.Sprintf("%s.Canon() = %s", tc.hand, tc.want), func(t *testing.T) {
			h0, err := parseHand(tc.hand)
			if err != nil {
				t.Fatalf("parseHand(%s) gave error %s", tc.hand, err)
			}
			N := len(h0)
			var h64 Hand64
			for _, c := range h0 {
				h64 = (h64 << 8) | Hand64(c)
			}
			got64c := h64.Canonical(N, 7)
			got := Hand64(got64c).String(len(h0))
			if got != tc.want {
				t.Errorf("%s.Canon() = %s, want %s", tc.hand, got, tc.want)
			}
			// We also check the examplar of the canonical form works.
			// We check it doesn't contain any x-suits, and that if we canonicalize
			// the exemplar, we get the same canonical form back.
			ex64 := got64c.Exemplar(N)
			ex64s := ex64.String(N)
			if strings.Contains(ex64s, "s") {
				t.Errorf("%s.Canon().Exemplar() = %s, still contains x-suit", tc.hand, ex64s)
			}
			rtCanon := ex64.Canonical(N, 7)
			if rtCanon != got64c {
				t.Errorf("%s.Canon().Exemplar().Canon() = %s, want %s", tc.hand, Hand64(rtCanon).String(N), got)
			}
		})
	}
}
