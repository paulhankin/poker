// +build ignore

package main

import (
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
	f, err := os.Create("tables_static.go")
	if err != nil {
		log.Fatalf("failed to create source file: %v", err)
	}
	tbl3, tbl5, tbl7 := poker.InternalTables()
	var werr error
	wf := func(v string, args ...interface{}) {
		if werr != nil {
			return
		}
		_, werr = fmt.Fprintf(f, v, args...)
	}
	wf(`
// +build staticdata

package poker

var rootNode7table  = [163060 * 52]uint32{`)
	fmt.Println("writing 7 table")
	for i, v := range tbl7 {
		if i%16 == 0 {
			wf("\n\t")
		}
		wf("0x%x, ", v)
	}
	wf(`
}

var rootNode5table = [3459 * 52]uint32{`)
	fmt.Println("writing 5 table")
	for i, v := range tbl5 {
		if i%16 == 0 {
			wf("\n\t")
		}
		wf("0x%x, ", v)
	}
	wf(`
}

var rootNode3table = [16 * 16 * 16]int16{`)
	fmt.Println("writing 3 table")
	for i, v := range tbl3 {
		if i%16 == 0 {
			wf("\n\t")
		}
		wf("0x%x, ", v)
	}
	wf(`
}
`)

	if werr != nil {
		log.Fatalf("write error: %v", werr)
	}
	if werr = f.Close(); werr != nil {
		log.Fatalf("failed to close file: %v", werr)
	}
}

func main() {
	writeSource()
	writeFile()
}
