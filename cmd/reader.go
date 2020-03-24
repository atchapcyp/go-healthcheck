package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
)

func ReadCSVFrom(path string) []string {
	csvfile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	r := csv.NewReader(csvfile)
	var records []string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		records = append(records, record[0])
	}

	return records
}
