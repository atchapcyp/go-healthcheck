package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendReport(t *testing.T) {
	t.Run("should return status 200 OK", func(t *testing.T) {
		mockClient := &MockClient{TargetStatus: 200}
		var ws = WebStat{}
		actual := ws.sendReport("/random/url", mockClient)
		expect := http.StatusOK
		assert.Equal(t, expect, actual)
	})

	t.Run("should return status 404", func(t *testing.T) {
		mockClient := &MockClient{TargetStatus: 404}
		var ws = WebStat{}
		actual := ws.sendReport("/random/url", mockClient)
		expect := http.StatusNotFound
		assert.Equal(t, expect, actual)
	})
}

func TestSenderConsumer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("10000 request 100 sender", func(t *testing.T) {
		var wg sync.WaitGroup
		var ws = WebStat{wg: &wg}
		var urls []string
		requestCount := 10000
		for i := 0; i < requestCount; i++ {
			urls = append(urls, ts.URL)
		}

		queue := make(chan *http.Request)
		ws.requestComposer(urls, queue)
		for i := 0; i < 100; i++ {
			ws.wg.Add(1)
			go ws.requestSender(queue)
		}
		ws.wg.Wait()

		assert.Equal(t, requestCount, ws.totalCheck(), "Total response are not equal to number of request")
	})

	t.Run("100 request 100 sender", func(t *testing.T) {
		var wg sync.WaitGroup
		var ws = WebStat{wg: &wg}
		var urls []string
		requestCount := 100
		for i := 0; i < requestCount; i++ {
			urls = append(urls, ts.URL)
		}

		queue := make(chan *http.Request)
		ws.requestComposer(urls, queue)
		for i := 0; i < 100; i++ {
			ws.wg.Add(1)
			go ws.requestSender(queue)
		}
		ws.wg.Wait()

		assert.Equal(t, requestCount, ws.totalCheck(), "Total response are not equal to number of request")
	})
}

func TestSetAccessToken(t *testing.T) {
	t.Run("should have access token in webstat", func(t *testing.T) {
		mockClient := &MockClient{
			TargetStatus:    200,
			WithAccessToken: true,
		}
		var ws = WebStat{}
		err := ws.setAccToken(mockClient)
		assert.NoError(t, err)
		assert.NotZero(t, ws.AccessToken)
	})

	t.Run("connection terminated should not have access token", func(t *testing.T) {
		mockClient := &MockClient{
			TargetStatus:         302,
			WithAccessToken:      true,
			ConnectionTerminated: true,
		}
		var ws = WebStat{}
		ws.setAccToken(mockClient)
		assert.Zero(t, ws.AccessToken)
	})

	t.Run("Unautorized, should not have access token", func(t *testing.T) {
		mockClient := &MockClient{
			TargetStatus: 401,
		}
		var ws = WebStat{}
		err := ws.setAccToken(mockClient)
		assert.Error(t, err)
		assert.Zero(t, ws.AccessToken)
	})
}
