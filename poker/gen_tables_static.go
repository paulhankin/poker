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
	f, err := os.Create("poker.dat")
	if err != nil {
		log.Fatalf("failed to create data file: %v", err)
	}
	tbl3, tbl5, tbl7 := poker.InternalTables()
	if err := binary.Write(f, binary.LittleEndian, tbl7[:]); err != nil {
		log.Fatalf("failed to write data: %v", err)
	}
	if err := binary.Write(f, binary.LittleEndian, tbl5[:]); err != nil {
		log.Fatalf("failed to write data: %v", err)
	}
	if err := binary.Write(f, binary.LittleEndian, tbl3[:]); err != nil {
		log.Fatalf("failed to write data: %v", err)
	}
	if err := f.Close(); err != nil {
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
	var werr error
	wf := func(v string, args ...interface{}) {
		if werr != nil {
			return
		}
		_, werr = fmt.Fprintf(f, v, args...)
	}
	wf(`
// +build !gendata,!filedata

package poker

var pokerTableData = []uint8("`)
	e64 := base64.NewEncoder(base64.RawStdEncoding, f)
	zs := gzip.NewWriter(e64)
	fmt.Println("writing 7 table")
	if err := binary.Write(zs, binary.LittleEndian, tbl7[:]); err != nil {
		log.Fatalf("failed to write data: %v", err)
	}
	fmt.Println("writing 5 table")
	if err := binary.Write(zs, binary.LittleEndian, tbl5[:]); err != nil {
		log.Fatalf("failed to write data: %v", err)
	}
	fmt.Println("writing 3 table")
	if err := binary.Write(zs, binary.LittleEndian, tbl3[:]); err != nil {
		log.Fatalf("failed to write data: %v", err)
	}
	if err := zs.Close(); err != nil {
		log.Fatalf("failed to close gzip: %v", err)
	}
	if err := e64.Close(); err != nil {
		log.Fatalf("failed to close base64 encoder", err)
	}
	wf("\")\n")

	if werr != nil {
		log.Fatalf("write error: %v", werr)
	}
	if werr = f.Flush(); werr != nil {
		log.Fatalf("failed to flush data: %v", werr)
	}
	if werr = rf.Close(); werr != nil {
		log.Fatalf("failed to close file: %v", werr)
	}
}

func main() {
	writeSource()
	writeFile()
}
