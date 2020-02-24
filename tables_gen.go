package poker

var (
	rootNode7table [163060 * 52]uint32
	rootNode5table [3459 * 52]uint32
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

func init() {
	genTables(5, rootNode5table[:], rootNode5(), make([]bool, len(rootNode5table)))
	genTables(7, rootNode7table[:], rootNode7(), make([]bool, len(rootNode7table)))
}
