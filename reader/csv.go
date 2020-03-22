package reader

import (
	"encoding/csv"
	"io"
	"log"
	"os"
)

type CSVRecords struct {
	Records []Record
}

type Record struct {
	URL    string
	Params []string
	Method string
}

func ReadCSVFrom(path string) CSVRecords {
	csvfile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	r := csv.NewReader(csvfile)
	var crs CSVRecords
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		crs.Records = append(crs.Records, Record{
			URL: record[0],
		})
	}
	return crs
}
