package mockhttp

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
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

func MockOutput(body interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(body)
	}
}

func TestClientOutput(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	client := New(MockOutput(map[string]string{"foo": "bar"}),
		Output(buf),
		BasicAuth("foo", "bar"),
	)
	resp, err := client.POST("/", map[string]string{"hello": "world"})
	if err != nil {
		t.Errorf("expected no error; got %v", err)
		return
	}

	rx := regexp.MustCompile(`\s+`)
	content := rx.ReplaceAllString(string(buf.Bytes()), "")

	// -- test request -------------------------------------------------

	request := rx.ReplaceAllString(`
POST http://localhost/
Authorization: Basic Zm9vOmJhcg==

{
  "hello": "world"
}
`, "")

	if !strings.Contains(content, request) {
		t.Errorf("expected substring %v; got %v", request, content)
	}

	var v map[string]string
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		t.Errorf("expected no error; got %v", err)
		return
	}
	if v["foo"] != "bar" {
		t.Errorf("expected output to leave the body alone; got %v", v)
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

func TestTransportGET(t *testing.T) {
	expected := "hello world"
	var fn http.HandlerFunc = func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusTemporaryRedirect)
		io.WriteString(w, expected)
	}
	client := New(fn)
	resp, err := client.TransportGET("/")
	if err != nil {
		t.Errorf("expected nil err; got %v", err)
		return
	}

	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("expected %v; got %v", http.StatusTemporaryRedirect, resp.StatusCode)
		return
	}

	data, _ := ioutil.ReadAll(resp.Body)
	if v := string(data); v != expected {
		t.Errorf("expected %v; got %v", expected, v)
		return
	}

	if v := resp.Header.Get("Location"); v != "/" {
		t.Errorf("expected %v; got %v", "/", v)
		return
	}
}
