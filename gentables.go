package poker

import (
	"fmt"
	"runtime"
	"sync"
)

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

type genwork struct {
	n    int
	h    Hand64Canonical
	node **Node
}

type genner struct {
	m     sync.Mutex
	cache map[Hand64Canonical]*Node
	work  chan genwork
	wg    sync.WaitGroup
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

func (g *genner) get(key Hand64Canonical) (*Node, bool) {
	g.m.Lock()
	defer g.m.Unlock()
	n, ok := g.cache[key]
	if ok {
		return n, true
	}
	n = &Node{}
	g.cache[key] = n
	return n, false
}

func (g *genner) genworker() {
	for w := range g.work {
		h := w.h
		n := w.n
		key := h | (Hand64Canonical(n) << (64 - 8))
		node, ok := g.get(key)
		*w.node = node
		if ok {
			g.wg.Add(-1)
			continue
		}
		fmt.Println(key)
		node.N = n
		node.H = h
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
				g.wg.Add(1)
				var nodeChild *Node
				g.work <- genwork{n: n + 1, h: nhc, node: &nodeChild}
				node.T[c] = Transition{
					SX: xf,
					N:  nodeChild,
				}
			}
		}
		g.wg.Add(-1)
	}
}

func Tree() *Node {
	g := &genner{
		cache: map[Hand64Canonical]*Node{},
		work:  make(chan genwork, 10000000),
	}
	g.wg.Add(1)
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			g.genworker()
		}()
	}
	node := &Node{}
	g.work <- genwork{node: &node}
	g.wg.Wait()
	close(g.work)
	return node
}
