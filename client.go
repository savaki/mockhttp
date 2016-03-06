package mockhttp

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"io/ioutil"
	"mime/multipart"
	"net/textproto"
)

type Client struct {
	handler http.Handler
	cookies map[string]string
}

type KV struct {
	Key   string
	Value string
}

func New(handler http.Handler) *Client {
	return &Client{
		handler: handler,
		cookies: map[string]string{},
	}
}

func (c *Client) Cookie(name string) (string, bool) {
	return c.cookies[name]
}

func (c *Client) DO(method, path string, header http.Header, body interface{}, keyvals ...KV) (*http.Response, error) {
	if c.cookies == nil {
		c.cookies = map[string]string{}
	}

	values := url.Values{}
	for _, kv := range keyvals {
		values.Add(kv.Key, kv.Value)
	}

	urlStr := "http://localhost" + path
	if len(values) > 0 {
		urlStr = urlStr + "?" + values.Encode()
	}

	var r io.Reader
	if body != nil {
		switch v := body.(type) {
		case []byte:
			r = bytes.NewReader(v)
		case io.Reader:
			r = v
		default:
			data, _ := json.Marshal(body)
			r = bytes.NewReader(data)
		}
	}

	// ---------------------------------------------
	// create the request
	//
	req, err := http.NewRequest(method, urlStr, r)
	if err != nil {
		log.Fatalln(err)
	}

	if header != nil {
		for key, values := range header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	cookies := ""
	for k, v := range c.cookies {
		cookies = cookies + "; " + k + "=" + v
	}
	if len(cookies) > 0 {
		cookies = cookies[2:]
	}
	req.Header.Set("Cookie", cookies)

	w := httptest.NewRecorder()

	// ---------------------------------------------
	// execute the handler
	//
	c.handler.ServeHTTP(w, req)

	// ---------------------------------------------
	// handle cookies in the response
	//

	// capture cookies
	if setCookie := w.Header().Get("Set-Cookie"); setCookie != "" {
		parts := strings.Split(setCookie, "=")
		name := parts[0]
		value := strings.Split(parts[1], ";")[0]
		c.cookies[name] = value
	}

	return &http.Response{
		StatusCode: w.Code,
		Request:    req,
		Header:     w.HeaderMap,
		Body:       ioutil.NopCloser(bytes.NewReader(w.Body.Bytes())),
	}, nil
}

func (c *Client) GET(path string, keyvals ...KV) (*http.Response, error) {
	return c.DO("GET", path, nil, nil, keyvals...)
}

func (c *Client) POST(path string, body interface{}) (*http.Response, error) {
	return c.DO("POST", path, nil, body)
}

func (c *Client) PUT(path string, body interface{}) (*http.Response, error) {
	return c.DO("PUT", path, nil, body)
}

func (c *Client) PATCH(path string, body interface{}) (*http.Response, error) {
	return c.DO("PATCH", path, nil, body)
}

func (c *Client) DELETE(path string, keyvals ...KV) (*http.Response, error) {
	return c.DO("DELETE", path, nil, nil, keyvals...)
}

func (c *Client) Upload(path string, r io.Reader) (*http.Response, error) {
	buf := bytes.NewBuffer([]byte{})

	m := multipart.NewWriter(buf)
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", `form-data; name="image"; filename="sample.png"`)
	h.Set("Content-Type", "image/png")
	w, _ := m.CreatePart(h)
	io.Copy(w, r)
	m.Close()

	header := http.Header{}
	header.Set("Content-Type", m.FormDataContentType())

	return c.DO("POST", path, header, bytes.NewReader(buf.Bytes()))
}
