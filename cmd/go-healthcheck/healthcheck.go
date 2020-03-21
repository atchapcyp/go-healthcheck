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

	var wg sync.WaitGroup
	var stat = WebStat{wg: &wg}
	fmt.Println("Perform website checking...")
	begin := time.Now()
	for _, r := range rc.Records {
		stat.wg.Add(1)
		go stat.webCheck(r.URL)
	}
	stat.wg.Wait()
	stat.totalTime = time.Since(begin)

	fmt.Println("Done!!")
	fmt.Printf("Checked website: %d\n", stat.Complete+stat.Failed)
	fmt.Printf("Successful websites : %d\n", stat.Complete)
	fmt.Printf("Failure websites : %d\n", stat.Failed)
	fmt.Println("Total times to finished checking website: ", stat.totalTime.Seconds())
}

type WebStat struct {
	Complete  int
	Failed    int
	wg        *sync.WaitGroup
	totalTime time.Duration
}

func (ws *WebStat) webCheck(url string) {
	defer ws.wg.Done()
	fmt.Println("url -> ", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: webCheck %v\n", err)
		ws.Failed++
	}
	if resp != nil {
		fmt.Println(resp.Status)
		ws.Complete++
	}
}
