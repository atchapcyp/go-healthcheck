package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var webList []string

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Require path to csv file")
		os.Exit(1)
	}

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func(c <-chan os.Signal) {
		<-c
		os.Exit(0)
	}(termChan)

	csvfile, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	r := csv.NewReader(csvfile)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		webList = append(webList, record[0])
	}
	fmt.Println(webList)

	// function Timer
	defer func(begin time.Time) {
		fmt.Printf("use %v", time.Since(begin).Seconds())
	}(time.Now())
	wg := &sync.WaitGroup{}
	for _, w := range webList {
		wg.Add(1)
		go webCheck(w, wg)
	}
	wg.Wait()
}

type webStat struct {
}

func webCheck(url string, wg *sync.WaitGroup) {
	fmt.Println("url -> ", url)
	_, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: webCheck %v\n", err)
	}
	wg.Done()
}
