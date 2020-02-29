// +build filedata

package poker

import (
	"compress/gzip"
	"encoding/binary"
	"os"
)

var (
	rootNode7table [163060 * 52]uint32
	rootNode5table [3459 * 52]uint32
	rootNode3table [16 * 16 * 16]int16
)

func init() {
	rf, err := os.Open("poker.dat")
	if err != nil {
		panic(err)
	}
	zf, err := gzip.NewReader(rf)
	if err != nil {
		panic(err)
	}
	if err := binary.Read(zf, binary.LittleEndian, rootNode7table[:]); err != nil {
		panic(err)
	}
	if err := binary.Read(zf, binary.LittleEndian, rootNode5table[:]); err != nil {
		panic(err)
	}
	if err := binary.Read(zf, binary.LittleEndian, rootNode3table[:]); err != nil {
		panic(err)
	}
	if err := zf.Close(); err != nil {
		panic(err)
	}
	if err := rf.Close(); err != nil {
		panic(err)
	}
}
