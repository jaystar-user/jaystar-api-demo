package httpUtil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type HttpClient struct {
	client  *http.Client
	timeout time.Duration
}

func ProvideHttpClient(timeout time.Duration) *HttpClient {
	return &HttpClient{
		timeout: timeout,
	}
}

func (c *HttpClient) Get(url string, query map[string]string, headers map[string]string) (int, []byte, error) {
	_url := fmt.Sprintf("%s?%s", url, queryMapEncode(query))
	return c.Request(http.MethodGet, _url, nil, headers)
}

func (c *HttpClient) Post(url string, body []byte, headers map[string]string) (int, []byte, error) {
	return c.Request(http.MethodPost, url, body, headers)
}

func (c *HttpClient) Put(url string, body []byte, headers map[string]string) (int, []byte, error) {
	return c.Request(http.MethodPut, url, body, headers)
}

func (c *HttpClient) Request(method string, url string, body []byte, headers map[string]string) (int, []byte, error) {
	c.client = &http.Client{
		Timeout: c.timeout,
	}

	var bodyReader io.Reader

	if len(body) > 0 {
		bodyReader = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return res.StatusCode, nil, err
	}

	readRes, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, nil, err
	}

	return res.StatusCode, readRes, nil
}

func queryMapEncode(qm map[string]string) string {
	uq := url.Values{}
	for k, v := range qm {
		if v != "" {
			uq.Add(k, v)
		}
	}
	return uq.Encode()
}
