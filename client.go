package mockhttp

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"regexp"
	"strings"
)

type Client struct {
	codebase string
	client   *http.Client
	authFunc func(*http.Request) error
}

type KV struct {
	Key   string
	Value string
}

func New(handler http.Handler, configs ...func(*Client)) *Client {
	cookieJar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar: cookieJar,
	}

	if handler != nil {
		httpClient.Transport = &roundTripper{handler: handler}
	}

	client := &Client{
		client: httpClient,
	}

	for _, config := range configs {
		config(client)
	}

	if client.codebase == "" {
		client.codebase = "http://localhost"
	}

	// strip trailing slashes from codebase
	pat := regexp.MustCompile(`/+$`)
	client.codebase = pat.ReplaceAllString(client.codebase, "")

	return client
}

func Codebase(codebase string) func(c *Client) {
	return func(c *Client) {
		c.codebase = codebase
	}
}

func BasicAuth(username, password string) func(c *Client) {
	return func(c *Client) {
		c.authFunc = func(req *http.Request) error {
			req.SetBasicAuth(username, password)
			return nil
		}
	}
}

func AuthFunc(authFunc func(*http.Request) error) func(c *Client) {
	return func(c *Client) {
		c.authFunc = authFunc
	}
}

func (c *Client) Cookie(name string) (string, bool) {
	u, err := url.Parse(c.codebase)
	if err != nil {
		return "", false
	}

	if cookies := c.client.Jar.Cookies(u); cookies != nil {
		for _, cookie := range cookies {
			if cookie.Name == name {
				return cookie.Value, true
			}
		}
	}

	return "", false
}

func (c *Client) DO(method, path string, header http.Header, body interface{}, keyvals ...KV) (*http.Response, error) {
	values := url.Values{}
	for _, kv := range keyvals {
		values.Add(kv.Key, kv.Value)
	}

	var urlStr string
	if strings.HasPrefix(path, "http") {
		urlStr = path
	} else {
		urlStr = c.codebase + path
	}
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
		return nil, err
	}

	if header != nil {
		for key, values := range header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	// handle authentication
	if c.authFunc != nil {
		if err = c.authFunc(req); err != nil {
			return nil, err
		}
	}

	return c.client.Do(req)
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
