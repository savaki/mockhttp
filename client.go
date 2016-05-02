package mockhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	codebase string
	client   *http.Client
	notifier Notifier
	authFunc func(*http.Request) error
	w        io.Writer
}

type KV struct {
	Key   string
	Value string
}

type Option func(*Client)

func New(handler http.Handler, options ...Option) *Client {
	cookieJar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar: cookieJar,
	}

	if handler != nil {
		httpClient.Transport = &roundTripper{handler: handler}
	}

	client := &Client{
		client:   httpClient,
		notifier: NopNotifier{},
		authFunc: func(*http.Request) error { return nil },
	}

	for _, opt := range options {
		opt(client)
	}

	if client.codebase == "" {
		client.codebase = "http://localhost"
	}

	// strip trailing slashes from codebase
	pat := regexp.MustCompile(`/+$`)
	client.codebase = pat.ReplaceAllString(client.codebase, "")

	return client
}

// Codebase allows one to specify a remote codebase (defaults to http://localhost)
func Codebase(codebase string) Option {
	return func(c *Client) {
		c.codebase = codebase
	}
}

// BasicAuth provides basic auth
func BasicAuth(username, password string) Option {
	return func(c *Client) {
		c.authFunc = func(req *http.Request) error {
			req.SetBasicAuth(username, password)
			return nil
		}
	}
}

// AuthFunc allows for an arbitrary authentication function
func AuthFunc(authFunc func(*http.Request) error) Option {
	return func(c *Client) {
		if authFunc == nil {
			c.authFunc = func(*http.Request) error { return nil }
		} else {
			c.authFunc = authFunc
		}
	}
}

// Output enables debug output to be written to the specified writer
func Output(w io.Writer) Option {
	return func(c *Client) {
		c.w = w
	}
}

func Observer(n Notifier) Option {
	return func(c *Client) {
		c.notifier = n
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

func newReader(body interface{}) io.Reader {
	if body == nil {
		return nil
	}

	var r io.Reader
	switch v := body.(type) {
	case []byte:
		r = bytes.NewReader(v)
	case io.Reader:
		r = v
	default:
		data, _ := json.Marshal(body)
		r = bytes.NewReader(data)
	}

	return r
}

func prettyReader(body interface{}) io.Reader {
	if body == nil {
		return nil
	}

	var r io.Reader
	switch body.(type) {
	case []byte:
		r = strings.NewReader("[binary content]")
	case io.Reader:
		r = strings.NewReader("[binary content]")
	default:
		data, _ := json.MarshalIndent(body, "", "  ")
		r = bytes.NewReader(data)
	}

	return r
}

func (c *Client) urlStr(path string, keyvals ...KV) string {
	values := url.Values{}
	for _, kv := range keyvals {
		values.Add(kv.Key, kv.Value)
	}

	var urlStr string
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		urlStr = path
	} else {
		urlStr = c.codebase + path
	}
	if len(values) > 0 {
		urlStr = urlStr + "?" + values.Encode()
	}

	return urlStr
}

func (c *Client) TransportGET(path string, keyvals ...KV) (*http.Response, error) {
	urlStr := c.urlStr(path, keyvals...)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	// handle authentication
	if err = c.authFunc(req); err != nil {
		return nil, err
	}

	roundTripper := c.client.Transport
	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}
	return roundTripper.RoundTrip(req)
}

func (c *Client) DO(method, path string, header http.Header, body interface{}, keyvals ...KV) (*http.Response, error) {
	urlStr := c.urlStr(path, keyvals...)
	r := newReader(body)

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
	if err = c.authFunc(req); err != nil {
		return nil, err
	}

	if c.w != nil {
		buf := bytes.NewBuffer([]byte{})
		fmt.Fprintf(buf, "\n#-- Request ------------------------------------------\n")
		fmt.Fprintf(buf, "%v %v\n", method, urlStr)
		for key, values := range req.Header {
			for _, value := range values {
				fmt.Fprintf(buf, "%v: %v\n", key, value)
			}
		}
		if body != nil {
			io.WriteString(buf, "\n")
			io.Copy(buf, prettyReader(body))
		}

		io.Copy(c.w, buf)
	}

	since := time.Now()
	resp, err := c.client.Do(req)
	if resp != nil {
		c.notifier.Notify(resp.StatusCode, req.Method, req.URL.Path, time.Now().Sub(since))
	}
	if err != nil {
		return nil, err
	}

	if c.w != nil {
		buf := bytes.NewBuffer([]byte{})

		if resp.Body != nil {
			defer resp.Body.Close()
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = ioutil.NopCloser(bytes.NewReader(data))

			fmt.Fprintf(buf, "\n\n#-- Response -----------------------------------------\n")
			fmt.Fprintf(buf, "%v\n", resp.Status)
			for key, values := range resp.Header {
				for _, value := range values {
					fmt.Fprintf(buf, "%v: %v\n", key, value)
				}
			}
			buf.Write(data)
		}

		fmt.Fprintf(buf, "\n\n#-- End ----------------------------------------------\n")
		io.Copy(c.w, buf)
	}

	return resp, nil
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
