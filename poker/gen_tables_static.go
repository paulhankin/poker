// +build ignore

package main

import (
	"bufio"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/paulhankin/poker/v2/poker"
)

func writeFile() {
	rf, err := os.Create("poker.dat")
	if err != nil {
		log.Fatalf("failed to create data file: %v", err)
	}
	zf := gzip.NewWriter(rf)
	tbl3, tbl5, tbl7 := poker.InternalTables()
	if err := binary.Write(zf, binary.LittleEndian, tbl7[:]); err != nil {
		log.Fatalf("failed to write data: %v", err)
	}
	if err := binary.Write(zf, binary.LittleEndian, tbl5[:]); err != nil {
		log.Fatalf("failed to write data: %v", err)
	}
	if err := binary.Write(zf, binary.LittleEndian, tbl3[:]); err != nil {
		log.Fatalf("failed to write data: %v", err)
	}
	if err := zf.Close(); err != nil {
		log.Fatalf("failed to write data: %v", err)
	}
	if err := rf.Close(); err != nil {
		log.Fatalf("failed to write data: %v", err)
	}
}

func writeSource() {
	rf, err := os.Create("tables_static.go")
	if err != nil {
		log.Fatalf("failed to create source file: %v", err)
	}
	f := bufio.NewWriter(rf)
	tbl3, tbl5, tbl7 := poker.InternalTables()
	if _, err := fmt.Fprint(f, `
// +build !gendata,!filedata

package poker

var pokerTableData = []uint8("`); err != nil {
		log.Fatal(err)
	}
	e64 := base64.NewEncoder(base64.RawStdEncoding, f)
	zs := gzip.NewWriter(e64)

	// We shrink the indexes for non-terminal/non-nil nodes.
	// We divide the index part by 52 (it's always a multiple),
	// and subtract the ID of the current node.
	// That reduces the compressed form from ~9MB to ~7.7MB.
	norm := func(tbl []uint32, n int) {
		for i := range tbl[:n*52] {
			if tbl[i] == 0 {
				continue
			}
			nodeIndex := uint32(i / 52)
			sx := tbl[i] & 0xff
			ix := (tbl[i] >> 8) / 52
			tbl[i] = sx | ((ix - nodeIndex) << 8)
		}
	}

	fmt.Println("writing 7 table")
	norm(tbl7, 61153)

	if err := binary.Write(zs, binary.LittleEndian, tbl7[:]); err != nil {
		log.Fatal(err)
	}
	fmt.Println("writing 5 table")
	norm(tbl5, 924)
	if err := binary.Write(zs, binary.LittleEndian, tbl5[:]); err != nil {
		log.Fatal(err)
	}
	fmt.Println("writing 3 table")
	if err := binary.Write(zs, binary.LittleEndian, tbl3[:]); err != nil {
		log.Fatal(err)
	}
	if err := zs.Close(); err != nil {
		log.Fatalf("failed to close gzip: %v", err)
	}
	if err := e64.Close(); err != nil {
		log.Fatalf("failed to close base64 encoder", err)
	}
	if _, err := fmt.Fprint(f, "\")\n"); err != nil {
		log.Fatal(err)
	}
	if err := f.Flush(); err != nil {
		log.Fatalf("failed to flush data: %v", err)
	}
	if err := rf.Close(); err != nil {
		log.Fatalf("failed to close file: %v", err)
	}
}

func main() {
	// Note! writeFile must come first, because writeSource overwrites
	// the data in the table.
	writeFile()
	writeSource()
}
