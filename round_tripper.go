package mockhttp

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

type roundTripper struct {
	handler http.Handler
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	r.handler.ServeHTTP(w, req)

	resp := &http.Response{
		StatusCode: w.Code,
		Request:    req,
		Header:     w.HeaderMap,
	}

	if w.Body != nil {
		resp.Body = ioutil.NopCloser(bytes.NewReader(w.Body.Bytes()))
	}

	return resp, nil
}
