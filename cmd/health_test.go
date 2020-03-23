package main

import (
	"net/http"
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

func TestWebRequest(t *testing.T) {
	t.Run("any response code should increment completed site", func(t *testing.T) {
		var wg sync.WaitGroup
		var ws = WebStat{wg: &wg}

		var httpStatusList = []int{101, 200, 201, 202, 204, 400, 401, 402, 403, 404, 500, 501, 502}
		for i, s := range httpStatusList {
			mockClient := &MockClient{
				TargetStatus: s,
			}
			ws.wg.Add(1)
			ws.webCheck("random/url", mockClient)
			assert.Equal(t, i+1, ws.Complete)
			assert.Equal(t, i+1, ws.totalCheck())
		}
	})

	t.Run("should increase failed site", func(t *testing.T) {
		var wg sync.WaitGroup
		var ws = WebStat{wg: &wg}
		var httpStatusList = []int{101, 200, 201, 202, 204, 400, 401, 402, 403, 404, 500, 501, 502}
		for i, s := range httpStatusList {
			mockClient := &MockClient{
				TargetStatus:         s,
				ConnectionTerminated: true, // terminate every connection.
			}
			ws.wg.Add(1)
			ws.webCheck("random/url", mockClient)
			assert.Equal(t, i+1, ws.Failed)
			assert.Equal(t, i+1, ws.totalCheck())
		}
	})
}
