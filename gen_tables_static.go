// +build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/paulhankin/poker"
)

func main() {
	f, err := os.Create("tables_static.go")
	if err != nil {
		log.Fatalf("failed to create target file %q: %v", os.Args[1], err)
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
// +build !gendata

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
