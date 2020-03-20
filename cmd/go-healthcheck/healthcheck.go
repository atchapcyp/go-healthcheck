package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/atchapcyp/go-healthcheck/reader"
)

var webList []string

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Require path to csv file")
		os.Exit(1)
	}

	terminateChan := make(chan os.Signal)
	signal.Notify(terminateChan, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func(c <-chan os.Signal) {
		<-c
		os.Exit(0)
	}(terminateChan)

	rc := reader.ReadCSVFrom(os.Args[1])
	// function Timer
	begin := time.Now()
	var wg sync.WaitGroup
	fmt.Println("Perform website checking...")
	for _, r := range rc.Records {
		wg.Add(1)
		go webCheck(r.URL, &wg)
	}
	wg.Wait()
	fmt.Println("Done!!")
	done := time.Since(begin).Seconds()
	fmt.Println("Checked website: ", len(webList))
	fmt.Println("Total times to finished checking website: ", done)
}

type webStat struct {
}

func webCheck(url string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("url -> ", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: webCheck %v\n", err)
	}
	if resp != nil {
		fmt.Println(resp.Status)
	}
}
