package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/atchapcyp/go-healthcheck/reader"
)

var (
	webList        []string
	TokenAccessURL = "https://api.line.me/oauth2/v2.1/token"
)

func main() {
	var reportURL, filePath string
	flag.StringVar(&reportURL, "u", "https://backend-challenge.line-apps.com/healthcheck/report", "Report target URL")
	flag.StringVar(&filePath, "f", "test.csv", "Path to CSV file")
	flag.Parse()

	terminateChan := make(chan os.Signal)
	signal.Notify(terminateChan, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func(c <-chan os.Signal) {
		<-c
		os.Exit(0)
	}(terminateChan)

	rc := reader.ReadCSVFrom(filePath)

	var wg sync.WaitGroup
	var stat = WebStat{wg: &wg}
	fmt.Println("Perform website checking...")
	go stat.setAccToken()
	stat.wg.Add(1)
	begin := time.Now()
	for _, r := range rc.Records {
		stat.wg.Add(1)
		go stat.webCheck(r.URL)
	}
	stat.wg.Wait()
	stat.totalTime = time.Since(begin)

	fmt.Println("Done!!")
	stat.printReport()
	stat.SendReport(reportURL)
}

type WebStat struct {
	Complete    int
	Failed      int
	wg          *sync.WaitGroup
	totalTime   time.Duration
	AccessToken string
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiredIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

func (ws *WebStat) webCheck(url string) {
	defer ws.wg.Done()
	defer func(begin time.Time) {
		fmt.Println(url, " Done in : ", time.Since(begin).Seconds())
	}(time.Now())

	tp := &http.Transport{
		DisableKeepAlives: true,
	}
	var client = http.Client{Transport: tp}
	resp, err := client.Get(url)
	if err != nil {
		ws.Failed++
	}
	if resp != nil {
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

func (ws *WebStat) SendReport(url string) {
	defer func(begin time.Time) {
		fmt.Println("SendReport Done in : ", time.Since(begin).Seconds())
	}(time.Now())

	reqBody, _ := json.Marshal(map[string]interface{}{
		"total_websites": ws.totalCheck(),
		"success":        ws.Complete,
		"failure":        ws.Failed,
		"total_time":     ws.totalTime.Nanoseconds(),
	})
	var client http.Client
	request, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", ws.AccessToken)
	resp, err := client.Do(request)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	fmt.Println("send report status", resp.Status)
}

func (ws *WebStat) setAccToken() {
	defer ws.wg.Done()
	var data = url.Values{}
	data.Add("grant_type", "refresh_token")
	data.Add("refresh_token", "ae6e3myyJpEOK6IkQDB6")
	data.Add("redirect_uro", "https://line-login-starter-20200321.herokuapp.com/auth")
	data.Add("client_id", "1653974782")
	data.Add("client_secret", "6830866844101ec965f282daca7b8808")
	var client http.Client
	request, err := http.NewRequest(http.MethodPost, TokenAccessURL, strings.NewReader(data.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(request)
	if err != nil {
		log.Println("unable to request for access token: ", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("unable to read access token response: ", err)
	}

	var tokenResponse TokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		log.Println("unable to unmarshal for access token", err)
	}

	ws.AccessToken = tokenResponse.AccessToken
}
