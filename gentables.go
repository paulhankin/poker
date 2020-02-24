package poker

import (
	"fmt"
	"runtime"
	"sync"
)

type tblTransition struct {
	rank int16 // for terminal nodes
	// SX describes how subsequent cards should
	// be transformed before
	SX SuitTransform
	N  *tblNode // The node to transform to
}

type tblNode struct {
	Index int
	N     int // number of cards
	H     hand64Canonical
	T     [52]tblTransition
}

type genwork struct {
	n    int
	h    hand64Canonical
	node **tblNode
}

type genner struct {
	m     sync.Mutex
	cache map[hand64Canonical]*tblNode
	work  chan genwork
	wg    sync.WaitGroup
}

func (g *genner) get(key hand64Canonical) (*tblNode, bool) {
	g.m.Lock()
	defer g.m.Unlock()
	n, ok := g.cache[key]
	if ok {
		return n, true
	}
	n = &tblNode{Index: len(g.cache)}
	g.cache[key] = n
	return n, false
}

func nodeeval5idx(c *[7]Card, idx [5]int) int16 {
	h := [5]Card{c[idx[0]], c[idx[1]], c[idx[2]], c[idx[3]], c[idx[4]]}
	return Eval5(&h)
}

func gentreeEval7(c *[7]Card) int16 {
	idx := [5]int{4, 3, 2, 1, 0}
	var best int16
	for {
		if ev := nodeeval5idx(c, idx); ev > best {
			best = ev
		}
		if idx[0] < 6 {
			idx[0]++
		} else if idx[1] < 5 {
			idx[1]++
			idx[0] = idx[1] + 1
		} else if idx[2] < 4 {
			idx[2]++
			idx[1] = idx[2] + 1
			idx[0] = idx[1] + 1
		} else if idx[3] < 3 {
			idx[3]++
			idx[2] = idx[3] + 1
			idx[1] = idx[2] + 1
			idx[0] = idx[1] + 1
		} else if idx[4] < 2 {
			idx[4]++
			idx[3] = idx[4] + 1
			idx[2] = idx[3] + 1
			idx[1] = idx[2] + 1
			idx[0] = idx[1] + 1
		} else {
			return best
		}
	}
}

func (g *genner) genworker(ncards int) {
	for w := range g.work {
		h := w.h
		n := w.n
		key := h | (hand64Canonical(n) << (64 - 8))
		node, ok := g.get(key)
		*w.node = node
		if ok {
			g.wg.Add(-1)
			continue
		}
		node.N = n
		node.H = h
		for c := 0; c < 52; c++ {
			nh, ok := h.Add(n, Card(c))
			if !ok {
				continue
			}
			nhc, xf := nh.CanonicalWithTransform(n+1, ncards)
			if n == ncards-1 {
				var rank int16
				if ncards == 7 {
					var c7 [7]Card
					copy(c7[:], nhc.Exemplar(7).CardsN(7))
					rank = gentreeEval7(&c7)
				} else if ncards == 5 {
					var c5 [5]Card
					copy(c5[:], nhc.Exemplar(5).CardsN(5))
					rank = EvalSlow(c5[:])
				} else {
					panic(ncards)
				}
				node.T[c] = tblTransition{
					rank: rank,
				}
			} else {
				node.T[c] = tblTransition{
					SX: xf,
				}
				g.wg.Add(1)
				work := genwork{n: n + 1, h: nhc, node: &node.T[c].N}
				go func() {
					g.work <- work
				}()
			}
		}
		g.wg.Add(-1)
	}
}

func gentree(ncards int) *tblNode {
	fmt.Println("generating tables for", ncards, "cards")

	g := &genner{
		cache: map[hand64Canonical]*tblNode{},
		work:  make(chan genwork, 10_000_000),
	}
	g.wg.Add(1)
	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			g.genworker(ncards)
			wg.Done()
		}()
	}
	node := &tblNode{}
	g.work <- genwork{node: &node}
	g.wg.Wait()
	close(g.work)
	wg.Wait()
	return node
}

var (
	rootNode5card     *tblNode
	rootNode5cardInit sync.Once

	rootNode7card     *tblNode
	rootNode7cardInit sync.Once
)

func rootNode7() *tblNode {
	rootNode7cardInit.Do(func() {
		rootNode7card = gentree(7)
	})
	return rootNode7card
}

func rootNode5() *tblNode {
	rootNode5cardInit.Do(func() {
		rootNode5card = gentree(5)
	})
	return rootNode5card
}

func nodeEval7(hand *[7]Card) int16 {
	node := rootNode7()
	tx := SuitTransform{0, 1, 2, 3}
	var t tblTransition
	for i := 0; i < 6; i++ {
		t = node.T[tx.Apply(hand[i])]
		tx = tx.Compose(t.SX)
		node = t.N
	}
	rank := node.T[tx.Apply(hand[6])].rank
	return rank
}

func nodeEval5(hand *[5]Card) int16 {
	node := rootNode5()
	tx := SuitTransform{0, 1, 2, 3}
	var t tblTransition
	for i := 0; i < 4; i++ {
		t = node.T[tx.Apply(hand[i])]
		tx = tx.Compose(t.SX)
		node = t.N
	}
	rank := node.T[tx.Apply(hand[4])].rank
	return rank
}

func Eval5(hand *[5]Card) int16 {
	idx := 0
	tx := SuitTransformByteIdentity
	var v uint32

	v = rootNode5table[idx+int(tx.Apply(hand[0]))]
	tx = tx.Compose(SuitTransformByte(v))
	idx = int(v >> 8)

	v = rootNode5table[idx+int(tx.Apply(hand[1]))]
	tx = tx.Compose(SuitTransformByte(v))
	idx = int(v >> 8)

	v = rootNode5table[idx+int(tx.Apply(hand[2]))]
	tx = tx.Compose(SuitTransformByte(v))
	idx = int(v >> 8)

	v = rootNode5table[idx+int(tx.Apply(hand[3]))]
	tx = tx.Compose(SuitTransformByte(v))
	idx = int(v >> 8)

	return int16(rootNode5table[idx+int(tx.Apply(hand[4]))])
}

func Eval7(hand *[7]Card) int16 {
	idx := 0
	tx := SuitTransformByteIdentity
	var v uint32

	v = rootNode7table[idx+int(tx.Apply(hand[0]))]
	tx = tx.Compose(SuitTransformByte(v))
	idx = int(v >> 8)

	v = rootNode7table[idx+int(tx.Apply(hand[1]))]
	tx = tx.Compose(SuitTransformByte(v))
	idx = int(v >> 8)

	v = rootNode7table[idx+int(tx.Apply(hand[2]))]
	tx = tx.Compose(SuitTransformByte(v))
	idx = int(v >> 8)

	v = rootNode7table[idx+int(tx.Apply(hand[3]))]
	tx = tx.Compose(SuitTransformByte(v))
	idx = int(v >> 8)

	v = rootNode7table[idx+int(tx.Apply(hand[4]))]
	tx = tx.Compose(SuitTransformByte(v))
	idx = int(v >> 8)

	v = rootNode7table[idx+int(tx.Apply(hand[5]))]
	tx = tx.Compose(SuitTransformByte(v))
	idx = int(v >> 8)

	return int16(rootNode7table[idx+int(tx.Apply(hand[6]))])
}
