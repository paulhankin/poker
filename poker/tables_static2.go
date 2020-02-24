// +build staticdata

package poker

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
)

var (
	rootNode7table [163060 * 52]uint32
	rootNode5table [3459 * 52]uint32
	rootNode3table [16 * 16 * 16]int16
)

func init() {
	rf := bytes.NewReader(pokerTableData)
	f, err := gzip.NewReader(rf)
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
}
