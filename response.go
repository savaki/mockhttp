package mockhttp

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
)

type Response interface {
	Code() int
	Reader() io.Reader
	UnmarshalJSON(v interface{}) error
	UnmarshalXML(v interface{}) error
	Header() http.Header
	Cookies() map[string]string
}

type response struct {
	w       *httptest.ResponseRecorder
	cookies map[string]string
}

func (r *response) Code() int {
	return r.w.Code
}

func (r *response) Reader() io.Reader {
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
