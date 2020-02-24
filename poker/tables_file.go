// +build !gendata,!staticdata

package poker

import (
	"encoding/binary"
	"os"
)

var (
	rootNode7table [163060 * 52]uint32
	rootNode5table [3459 * 52]uint32
	rootNode3table [16 * 16 * 16]int16
)

func init() {
	f, err := os.Open("poker.dat")
	if err != nil {
		panic(err)
	}
	if err := binary.Read(f, binary.LittleEndian, rootNode7table[:]); err != nil {
		panic(err)
	}
	if err := binary.Read(f, binary.LittleEndian, rootNode5table[:]); err != nil {
		panic(err)
	}
	if err := binary.Read(f, binary.LittleEndian, rootNode3table[:]); err != nil {
		panic(err)
	}
	f.Close()
}
