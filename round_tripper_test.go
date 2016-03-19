package mockhttp

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"testing"
)

func MockBody(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, body)
	}
}

func TestRoundTrip(t *testing.T) {
	body := "hello world"

	// Given
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Transport: &roundTripper{handler: MockBody(body)},
		Jar:       cookieJar,
	}

	resp, err := client.Get("/")
	if err != nil {
		t.Errorf("expected no err; got %v", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("expected no err; got %v", err)
	}
	if v := string(data); v != body {
		t.Errorf("expected body %v; got %v", body, v)
	}
}
