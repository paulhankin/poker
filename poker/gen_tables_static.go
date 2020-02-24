// +build ignore

package main

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/paulhankin/poker/v2/poker"
)

type byteWriter struct {
	f io.Writer
	n int
}

func (bw *byteWriter) Write(p []byte) (int, error) {
	i := bw.n
	for _, b := range p {
		if i%16 == 0 {
			if _, err := fmt.Fprintf(bw.f, "\n\t"); err != nil {
				return 0, err
			}
		}
		if _, err := fmt.Fprintf(bw.f, "0x%02x,", b); err != nil {
			return 0, err
		}
		i++
	}
	bw.n = i
	return len(p), nil
}

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
// +build staticdata

package poker

var pokerTableData = []uint8{`)
	zs := gzip.NewWriter(&byteWriter{f: f})
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
	wf("\n}\n")

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
