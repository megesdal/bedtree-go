package main

import (
	"encoding/csv"
	"github.com/megesdal/bedtree-go/bedtree"
	"io"
	"log"
	"os"
	"time"
)

func main() {

	start := time.Now()

	filename := os.Args[1]

	index := bedtree.NewIndex()

	nentries := populateIndex(filename, index)

	log.Printf("Index has %d unique keys across %d entries\n", index.Size(), nentries)
	log.Printf("FINISHED [%dms]\n", time.Now().Sub(start)/time.Millisecond)
}

func populateIndex(filename string, index *bedtree.Index) int {

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(file)
	csvReader.Read()

	count := 0
	for true {

		fields, err := csvReader.Read()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		count += 1
		index.Insert(fields[0])
	}

	return count
}
