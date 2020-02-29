// +build !gendata,!filedata

package poker

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
)

var (
	rootNode7table [163060 * 52]uint32
	rootNode5table [3459 * 52]uint32
	rootNode3table [16 * 16 * 16]int16
)

func denorm(tbl []uint32, n int) {
	for i := 0; i < n*52; i++ {
		if tbl[i] == 0 {
			continue
		}
		ni := uint32(i / 52)
		ix := tbl[i] >> 8
		tbl[i] = (tbl[i] & 0xff) | ((ix+ni)*52)<<8
	}
}

func init() {
	rf := bytes.NewReader(pokerTableData)
	d64f := base64.NewDecoder(base64.RawStdEncoding, rf)
	f, err := gzip.NewReader(d64f)
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
	if err := f.Close(); err != nil {
		panic(err)
	}

	denorm(rootNode5table[:], 924)
	denorm(rootNode7table[:], 61153)
}
