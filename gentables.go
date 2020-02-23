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
	Index int
	N     int // number of cards
	H     Hand64Canonical
	T     [52]Transition
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

func (g *genner) get(key Hand64Canonical) (*Node, bool) {
	g.m.Lock()
	defer g.m.Unlock()
	n, ok := g.cache[key]
	if ok {
		return n, true
	}
	n = &Node{Index: len(g.cache)}
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
		key := h | (Hand64Canonical(n) << (64 - 8))
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
					rank = EvalSlow5(&c5)
				} else {
					panic(ncards)
				}
				node.T[c] = Transition{
					rank: rank,
				}
			} else {
				node.T[c] = Transition{
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

func gentree(ncards int) *Node {
	fmt.Println("generating tables for", ncards, "cards")

	g := &genner{
		cache: map[Hand64Canonical]*Node{},
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
	node := &Node{}
	g.work <- genwork{node: &node}
	g.wg.Wait()
	close(g.work)
	wg.Wait()

	var indextable []uint32
	if ncards == 5 {
		indextable = rootNode5table[:]
	} else if ncards == 7 {
		indextable = rootNode7table[:]
	}
	for _, node := range g.cache {
		table := indextable[node.Index*52 : (node.Index+1)*52]
		if node.N == ncards-1 {
			for i, t := range node.T {
				table[i] = uint32(t.rank)
			}
		} else {
			for i, t := range node.T {
				if t.N != nil {
					table[i] = (uint32(t.N.Index*52) << 8) | uint32(t.SX.Byte())
				}
			}
		}
	}

	fmt.Println("nodes created for", ncards, "cards:", len(g.cache))
	return node
}

var (
	rootNode5card  *Node
	rootNode5table [3459 * 52]uint32

	rootNode7card     *Node
	rootNode7table    [163060 * 52]uint32
	rootNode7cardInit sync.Once
)

func rootNode7() *Node {
	rootNode7cardInit.Do(func() {
		rootNode7card = gentree(7)
	})
	return rootNode7card
}

func init() {
	rootNode5card = gentree(5)
}

func rootNode5() *Node {
	return rootNode5card
}

func init() {
	rootNode7()
}

func NodeEval7(hand *[7]Card) int16 {
	node := rootNode7()
	tx := SuitTransform{0, 1, 2, 3}
	var t Transition
	for i := 0; i < 6; i++ {
		t = node.T[tx.Apply(hand[i])]
		tx = tx.Compose(t.SX)
		node = t.N
	}
	rank := node.T[tx.Apply(hand[6])].rank
	return rank
}

func NodeEval5(hand *[5]Card) int16 {
	node := rootNode5()
	tx := SuitTransform{0, 1, 2, 3}
	var t Transition
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
