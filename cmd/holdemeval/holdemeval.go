// Binary holdemeval computes exact holdem hand equities for a given set
// of hands.
// For example:
//   holdemeval -hands "AcKh KdTh QhQd" -board 7d8c8sTs
// The board can be empty (in which case they are preflop equities),
// or any number of cards up to 5.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/paulhankin/poker"
)

var (
	handsFlag = flag.String("hands", "", "hands to compare")
	boardFlag = flag.String("board", "", "board cards to start with")
)

func parseCard(s string) (poker.Card, error) {
	if len(s) != 2 {
		return 0, fmt.Errorf("card should be of length 2, like Ac, but got %q", s)
	}
	c, ok := poker.NameToCard[strings.ToUpper(s)]
	if ok {
		return c, nil
	}
	// try with suit and rank the other way round
	c, ok = poker.NameToCard[strings.ToUpper(s[1:]+s[:1])]
	if ok {
		return c, nil
	}
	return 0, fmt.Errorf("failed to parse card %q", s)
}

func parseHand(s string) ([2]poker.Card, error) {
	var hz [2]poker.Card
	if len(s) != 4 {
		return hz, fmt.Errorf("expect hand in format like AcKh, got %q", s)
	}
	c0, err0 := parseCard(s[:2])
	c1, err1 := parseCard(s[2:])
	if err0 == nil {
		err0 = err1
	}
	if err0 != nil {
		return hz, err0
	}
	return [2]poker.Card{c0, c1}, nil
}

func fmtHand(h [2]poker.Card) string {
	s0 := h[0].Rank().String() + strings.ToLower(h[0].Suit().String())
	s1 := h[1].Rank().String() + strings.ToLower(h[1].Suit().String())
	return s0 + s1
}

func main() {
	flag.Parse()
	var hands [][2]poker.Card

	fail := func(err error) {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}

	if len(*handsFlag) == 0 {
		fail(fmt.Errorf("must specify one or more hands via the -hands flag"))
	}

	for _, p := range strings.Fields(*handsFlag) {
		h, err := parseHand(p)
		if err != nil {
			fail(err)
		}
		hands = append(hands, h)
	}

	brd := strings.ReplaceAll(*boardFlag, " ", "")
	if len(brd)%2 != 0 {
		fail(fmt.Errorf("bad -board flag %q. Missing a suit or rank?", *boardFlag))
	}
	var board []poker.Card
	for i := 0; i < len(brd); i += 2 {
		c, err := parseCard(brd[i : i+2])
		if err != nil {
			fail(fmt.Errorf("bad -board flag %q: %v", *boardFlag, err))
		}
		board = append(board, c)
	}

	eqs, err := poker.HoldemEquities(hands, board)
	if err != nil {
		fail(fmt.Errorf("failed to compute equities: %v", err))
	}
	fmt.Printf("%d runouts evaluated\n", eqs[0].Boards)
	for i := 0; i < len(hands); i++ {
		fmt.Printf("%s: equity:%.02f%%\twin:%.02f%%\ttie:%.02f%%\n", fmtHand(hands[i]), eqs[i].Equity*100, eqs[i].Win*100, eqs[i].Tie*100)
	}

}
