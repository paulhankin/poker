package poker

import "fmt"

type Transition struct {
	rank int16 // for terminal nodes
	// SX describes how subsequent cards should
	// be transformed before
	SX SuitTransform
	N  *Node // The node to transform to
}

type Node struct {
	N int // number of cards
	H Hand64Canonical
	T [52]Transition
}

func genTree(n int, h Hand64Canonical, cache map[Hand64Canonical]*Node) *Node {
	key := h | (Hand64Canonical(n) << (64 - 8))
	if cache[key] != nil {
		return cache[key]
	}
	fmt.Printf("%0*x\n", 2*n, h)
	node := Node{N: n, H: h}
	for c := 0; c < 52; c++ {
		nh, ok := h.Add(n, Card(c))
		if !ok {
			continue
		}
		nhc, xf := nh.CanonicalWithTransform(n)
		if n == 6 {
			var c7 [7]Card
			copy(c7[:], nhc.Exemplar(7).CardsN(7))
			rank := Eval7(&c7)
			node.T[c] = Transition{
				rank: rank,
			}
		} else {
			node.T[c] = Transition{
				SX: xf,
				N:  genTree(n+1, nhc, cache),
			}
		}
	}
	cache[key] = &node
	return &node
}

func Tree() *Node {
	cache := map[Hand64Canonical]*Node{}
	return genTree(0, 0, cache)
}
