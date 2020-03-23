package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
)

type MockClient struct {
	TargetStatus         int
	TargetResponse       string
	WithAccessToken      bool
	ConnectionTerminated bool
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	if m.ConnectionTerminated {
		return nil, errors.New("connection timed out..")
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(m.TargetStatus)
		w.Write([]byte(m.TargetResponse))
		if m.WithAccessToken {
			w.Write([]byte(`{"access_token":"eyJhbGciOiJIUzI1NiJ9.DexoW1tJGC8PFgPwah7ZlDQJhHb7j35LEj7AEt82kNU3sKdjGBYuFsJNDF2jgNr-4ecjtbKbvQOJwcXJ33LxE6OUgWbTDoEOcBHhZMLHsugvuaD51lgViMbbNm0j-Az8vQiVeZY5Y3MiFGh3175qPQmlga_Jpu5B0uFEmds38B4.eqKNmhg4c6sLEaujyijDzGmuepSpC_VVf9mrTSf0yQk","token_type":"Bearer","refresh_token":"ae6e3myyJpEOK6IkQDB6","expires_in":2592000,"scope":"profile openid"}`))
		}
	}))
	defer ts.Close()

	return http.Get(ts.URL)
}
