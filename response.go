package mockhttp

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
)

type Response interface {
	Request() *http.Request
	Code() int
	Body() io.Reader
	UnmarshalJSON(v interface{}) error
	UnmarshalXML(v interface{}) error
	Header() http.Header
	Cookies() map[string]string
}

type response struct {
	req     *http.Request
	w       *httptest.ResponseRecorder
	cookies map[string]string
}

func (r *response) Request() *http.Request {
	return r.req
}

func (r *response) Code() int {
	return r.w.Code
}

func (r *response) Body() io.Reader {
	return r.w.Body
}

func (r *response) UnmarshalJSON(v interface{}) error {
	return json.NewDecoder(r.w.Body).Decode(v)
}

func (r *response) UnmarshalXML(v interface{}) error {
	return xml.NewDecoder(r.w.Body).Decode(v)
}

func (r *response) Header() http.Header {
	return r.w.Header()
}

func (r *response) Cookies() map[string]string {
	return r.cookies
}
