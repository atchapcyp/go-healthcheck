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

	noFile := fileDescriptorSize()
	if noFile < 0 {
		noFile = 100
	}

	urls := readCSVFrom(filePath)

	var wg sync.WaitGroup
	var stat = WebStat{wg: &wg}

	queue := stat.requestComposer(urls)

	if err := stat.setAccToken(NewHttpClient()); err != nil {
		panic(err)
	}
	fmt.Println("Perform website checking...")

	begin := time.Now()
	for i := 0; i < 30; i++ {
		stat.wg.Add(1)
		go stat.requestSender(queue, i)
	}
	stat.wg.Wait()
	stat.totalTime = time.Since(begin)

	fmt.Println("Done!!")

	if status := stat.sendReport(reportURL, NewHttpClient()); status != 200 {
		log.Println("SendReport failed ... ", status)
		return
	}
	stat.printReport()
}

func NewHttpClient() *http.Client {
	return &http.Client{Transport: &http.Transport{
		DisableKeepAlives: true,
	}, Timeout: time.Duration(time.Second * 20)}
}

func (ws *WebStat) webCheck(url string, client HttpClienter) {
	defer ws.wg.Done()
	defer func(begin time.Time) {
		fmt.Println(url, " Done in : ", time.Since(begin).Seconds())
	}(time.Now())

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		ws.Failed++
		return
	}
	request.Header.Set("Connection", "close")
	request.Close = true

	resp, err := http.Get(url)
	if resp != nil {
		ws.Complete++
		resp.Body.Close()
		resp.Close = true
		// resp.Close = true
	}
	if err != nil {
		ws.Failed++
		fmt.Println("webCheck error : ", err)
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

func (ws *WebStat) requestComposer(urls []string) chan *http.Request {
	queue := make(chan *http.Request)

	go func() {
		for _, url := range urls {
			req, _ := http.NewRequest(http.MethodGet, url, nil)
			queue <- req
		}

		close(queue)
	}()

	return queue
}

func (ws *WebStat) requestSender(queue chan *http.Request, id int) {
	defer func() { ws.wg.Done() }()

	client := NewHttpClient()
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
				fmt.Println("requestSender error : ", err)
			}
			time.Sleep(time.Millisecond * 10)
			fmt.Printf("consumer No: %d, url: %s\n", id, string(req.URL.String()))
		}
	}

}
