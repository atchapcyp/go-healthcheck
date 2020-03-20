package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Require path to csv file")
		os.Exit(1)
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if readCSV(os.Args[1]) {
				fmt.Println("done ReadCSV")
			}
		}
	}
}

func readCSV(path string) bool {
	fmt.Println("path ", path)
	return true
}
