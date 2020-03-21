package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	stat.printReport()
	stat.SendReport()
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

func (ws *WebStat) printReport() {
	fmt.Printf("Checked website: %d\n", ws.totalCheck())
	fmt.Printf("Successful websites : %d\n", ws.Complete)
	fmt.Printf("Failure websites : %d\n", ws.Failed)
	fmt.Println("Total times to finished checking website: ", ws.totalTime.Seconds())
}

func (ws *WebStat) totalCheck() int {
	return ws.Complete + ws.Failed
}

func getLineToken() (string, error) {

	return "", nil
}

func (ws *WebStat) SendReport() {
	url := "https://backend-challenge.line-apps.com/healthcheck/report"
	accToken, err := getLineToken()
	if err != nil {
		log.Fatalln(err)
	}
	reqBody, err := json.Marshal(map[string]interface{}{
		"total_websites": ws.totalCheck(),
		"success":        ws.Complete,
		"failure":        ws.Failed,
		"total_time":     ws.totalTime.Nanoseconds(),
	})
	client := http.Client{}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", accToken)
	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	fmt.Printf("response : %+v\n", resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("body : ", body)
}
