package main

import (
	"net/http"
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
		ws.setAccToken(mockClient)
		assert.NotZero(t, ws.AccessToken)
	})

	t.Run("connection terminated should not have access token", func(t *testing.T) {
		mockClient := &MockClient{
			TargetStatus:         401,
			WithAccessToken:      true,
			ConnectionTerminated: true,
		}
		var ws = WebStat{}
		ws.setAccToken(mockClient)
		assert.Zero(t, ws.AccessToken)
	})

	t.Run("Unautorized should not have access token", func(t *testing.T) {
		mockClient := &MockClient{
			TargetStatus: 401,
		}
		var ws = WebStat{}
		ws.setAccToken(mockClient)
		assert.Zero(t, ws.AccessToken)
	})
}
