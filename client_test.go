package mockhttp

import (
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func MockCookie(cookieName string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie(cookieName)
		if err == http.ErrNoCookie {
			http.SetCookie(w, &http.Cookie{
				Name:  cookieName,
				Value: time.Now().String(),
				Path:  "/",
			})
			return
		}

		io.WriteString(w, cookie.Value)
	}
}

func TestClientIsCookieAware(t *testing.T) {
	// Given
	cookieName := "woot"
	client := New(MockCookie(cookieName))

	// When - make request returns cookie
	resp, err := client.GET("/")
	if err != nil {
		t.Errorf("expected nil err; got %v", err)
		return
	}
	if v := len(resp.Cookies()); v != 1 {
		t.Errorf("expected 1 cookie; got %v", v)
		return
	}
	cookieValue := resp.Cookies()[0].Value

	// Then - second request returns same cookie
	resp, err = client.GET("/")
	if err != nil {
		t.Errorf("expected nil err; got %v", err)
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("expected nil err; got %v", err)
		return
	}
	if v := string(data); v != cookieValue {
		t.Errorf("expected %v; got %v", cookieValue, v)
	}
}
