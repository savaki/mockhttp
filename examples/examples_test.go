package examples_test

import (
	"net/http"
	"testing"

	"io/ioutil"

	"github.com/savaki/mockhttp"
	"github.com/savaki/mockhttp/examples"
)

func TestNotFound(t *testing.T) {
	app := mockhttp.New(examples.Router())
	resp := app.GET("/invalid-path")
	if resp.Code() != http.StatusNotFound {
		t.Fail()
	}
}

func testPOST(t *testing.T) {
	app := mockhttp.New(examples.Router())
	resp := app.POST("/greeting", examples.GreetingIn{Name: "Matt"})
	if resp.Code() != http.StatusOK {
		t.Fail()
	}

	out := examples.GreetingOut{}
	err := resp.UnmarshalJSON(&out)
	if err != nil {
		t.Fail()
	}

	if out.Message != "Hello Matt" {
		t.Fail()
	}
}

func testGET(t *testing.T) {
	message := "argle-bargle"
	app := mockhttp.New(examples.Router())
	resp := app.GET("/echo", mockhttp.KV{"q", message})
	if resp.Code() != http.StatusOK {
		t.Fail()
	}

	data, err := ioutil.ReadAll(resp.Body())
	if err != nil {
		t.Fail()
	}
	if string(data) != message {
		t.Fail()
	}
}
