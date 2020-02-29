// +build gendata

package poker

var (
	rootNode7table [163060 * 52]uint32
	rootNode5table [3459 * 52]uint32
	rootNode3table [16 * 16 * 16]int16
)

func genTables(ncards int, indextable []uint32, node *tblNode, done []bool) {
	table := indextable[node.Index*52 : (node.Index+1)*52]
	if node.N == ncards-1 {
		for i, t := range node.T {
			table[i] = uint32(t.rank)
		}
		return
	}
	for i, t := range node.T {
		if t.N == nil {
			continue
		}
		table[i] = (uint32(t.N.Index*52) << 8) | uint32(t.SX.Byte())
		if !done[t.N.Index] {
			done[t.N.Index] = true
			genTables(ncards, indextable, t.N, done)
		}
	}
}

// The 3-card tables are simpler: we build a table with the rank for
// each triple of cards. Hand c1,c2,c3 is stored at index r1*256+r2*16+r3
// where r1, r2, r3 are the ranks (from 0 to 12) of the cards c1,c2,c3.
func genTables3(indextable []int16) {
	var cards [3]Card
	for i := 0; i < 13; i++ {
		cards[0], _ = MakeCard(Club, Rank(1+i))
		for j := 0; j < 13; j++ {
			cards[1], _ = MakeCard(Diamond, Rank(1+j))
			for k := 0; k < 13; k++ {
				cards[2], _ = MakeCard(Heart, Rank(1+k))
				indextable[i*256+j*16+k] = EvalSlow(cards[:])
			}
		}
	}
}

func init() {
	genTables(5, rootNode5table[:], rootNode5(), make([]bool, len(rootNode5table)))
	genTables(7, rootNode7table[:], rootNode7(), make([]bool, len(rootNode7table)))
	genTables3(rootNode3table[:])
}
