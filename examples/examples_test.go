package examples_test

import (
	"io/ioutil"
	"net/http"
	"testing"
	"encoding/json"

	"github.com/savaki/mockhttp"
	"github.com/savaki/mockhttp/examples"
)

func TestNotFound(t *testing.T) {
	app := mockhttp.New(examples.Router())
	resp, err := app.GET("/invalid-path")
	if err != nil {
		t.Fail()
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fail()
	}
}

func TestPOST(t *testing.T) {
	app := mockhttp.New(examples.Router())
	resp, err := app.POST("/greeting", examples.GreetingIn{Name: "Matt"})
	if err != nil {
		t.Fail()
	}
	if resp.StatusCode != http.StatusOK {
		t.Fail()
	}

	out := examples.GreetingOut{}
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		t.Fail()
	}

	if out.Message != "Hello Matt" {
		t.Fail()
	}
}

func TestGET(t *testing.T) {
	message := "argle-bargle"
	app := mockhttp.New(examples.Router())
	resp, err := app.GET("/echo", mockhttp.KV{"q", message})
	if err != nil {
		t.Fail()
	}
	if resp.StatusCode != http.StatusOK {
		t.Fail()
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fail()
	}
	if string(data) != message {
		t.Fail()
	}
}
