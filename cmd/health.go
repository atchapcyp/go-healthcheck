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

	"github.com/pkg/errors"
)

var (
	webList        []string
	TokenAccessURL = "https://api.line.me/oauth2/v2.1/token"
)

type WebStat struct {
	Complete    int
	Failed      int
	wg          *sync.WaitGroup
	totalTime   time.Duration
	AccessToken string
	timeout     int
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiredIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

type HttpClienter interface {
	Do(req *http.Request) (*http.Response, error)
}

func main() {
	var reportURL, filePath string
	var goroutineNum, requestTimeout int
	flag.StringVar(&reportURL, "u", "https://backend-challenge.line-apps.com/healthcheck/report", "Report target URL")
	flag.StringVar(&filePath, "f", "test.csv", "Path to CSV file")
	flag.IntVar(&goroutineNum, "n", 70, "A number of concurrent request.")
	flag.IntVar(&requestTimeout, "t", 30, "Request time out in second")
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

	// goroutineNum define the maximum number of the request consumer
	// goroutineNum should less than current file discriptor limit to prevent io error
	// WARNING : Using a too big `goroutineNum` means this program need to handle a large amount of goroutine.
	fs := fileDescriptorSize()
	if fs > 0 && fs < goroutineNum {
		goroutineNum = fs * 7 / 10
	}

	urls := readCSVFrom(filePath)

	var ws = WebStat{
		wg:      &sync.WaitGroup{},
		timeout: requestTimeout}

	if err := ws.setAccToken(NewHttpClient(requestTimeout)); err != nil {
		panic(err)
	}
	fmt.Println("Perform website checking...")

	queue := make(chan *http.Request)
	begin := time.Now()
	for i := 0; i < goroutineNum; i++ {
		ws.wg.Add(1)
		go ws.requestSender(queue)
	}
	ws.requestComposer(urls, queue)
	ws.wg.Wait()
	ws.totalTime = time.Since(begin)

	fmt.Println("Done!!")

	if status := ws.sendReport(reportURL, NewHttpClient(requestTimeout)); status != 200 {
		log.Println("SendReport failed ... ", status)
		return
	}
	ws.printReport()

	fmt.Println("goroutine num", goroutineNum)
}

func NewHttpClient(t int) *http.Client {
	return &http.Client{Transport: &http.Transport{
		DisableKeepAlives: true,
	}, Timeout: time.Duration(time.Duration(t) * time.Second), Jar: nil}
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

func (ws *WebStat) sendReport(url string, client HttpClienter) int {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"total_websites": ws.totalCheck(),
		"success":        ws.Complete,
		"failure":        ws.Failed,
		"total_time":     ws.totalTime.Nanoseconds(),
	})
	request, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", ws.AccessToken)
	resp, err := client.Do(request)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	return resp.StatusCode
}

func (ws *WebStat) setAccToken(client HttpClienter) error {
	var data = url.Values{}
	data.Add("grant_type", "refresh_token")
	data.Add("refresh_token", "ae6e3myyJpEOK6IkQDB6")
	data.Add("redirect_uro", "https://line-login-starter-20200321.herokuapp.com/auth")
	data.Add("client_id", "1653974782")
	data.Add("client_secret", "6830866844101ec965f282daca7b8808")
	request, err := http.NewRequest(http.MethodPost, TokenAccessURL, strings.NewReader(data.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var resp *http.Response
	if resp, err = client.Do(request); err != nil || resp == nil {
		return errors.Wrapf(err, "unable to get access token")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "unable to read access token response")
	}

	var tokenResponse TokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return errors.Wrapf(err, "unable to unmarshal for access token")
	}

	ws.AccessToken = tokenResponse.AccessToken
	return nil
}

func (ws *WebStat) requestComposer(urls []string, queue chan *http.Request) {
	go func() {
		for _, url := range urls {

			req, _ := http.NewRequest(http.MethodGet, url, nil)
			queue <- req
		}
		close(queue)
	}()
}

func (ws *WebStat) requestSender(queue chan *http.Request) {
	defer ws.wg.Done()

	client := NewHttpClient(ws.timeout)
	for {
		select {
		case req, ok := <-queue:
			if !ok {
				return
			}

			resp, err := client.Do(req)
			if resp != nil {
				ws.Complete++
				resp.Body.Close()
			}
			if err != nil {
				ws.Failed++
				log.Println("Error request to ", req.URL.String(), "err :", err)
			}
		}
	}
}
