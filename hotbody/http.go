package hotbody

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/http2"

	"boscoin.io/sebak/lib/errors"
)

type HTTP2Client struct {
	timeout   time.Duration
	url       *url.URL
	client    http.Client
	transport *http.Transport
	headers   http.Header
}

func NewHTTP2Client(timeout time.Duration, url *url.URL, headers http.Header) (http2Client *HTTP2Client, err error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		IdleConnTimeout:   timeout,
		DisableKeepAlives: true,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			DualStack: true,
		}).DialContext,
	}

	if err = http2.ConfigureTransport(transport); err != nil {
		return
	}

	client := http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // NOTE prevent redirect
		},
	}
	http2Client = &HTTP2Client{
		timeout:   timeout,
		url:       url,
		client:    client,
		transport: transport,
		headers:   headers,
	}

	return
}

func (client *HTTP2Client) URL() *url.URL {
	return client.url
}

func (client *HTTP2Client) Transport() *http.Transport {
	return client.transport
}

func (client *HTTP2Client) resolvePath(path string) *url.URL {
	return client.url.ResolveReference(&url.URL{Path: path})
}

func (client *HTTP2Client) newHeaders(headers http.Header) http.Header {
	newHeaders := http.Header{}
	for k, v := range client.headers {
		newHeaders[k] = v
	}

	if headers != nil {
		for k, v := range headers {
			newHeaders[k] = v
		}
	}

	return newHeaders
}

func (client *HTTP2Client) request(method, path string, body io.Reader, headers http.Header) (response *http.Response, err error) {
	u := client.resolvePath(path)

	var r *http.Request
	if r, err = http.NewRequest(method, u.String(), body); err != nil {
		return
	}
	defer func() {
		r.Close = true
	}()

	r.Header = client.newHeaders(headers)

	if client.timeout > 0 {
		ctx, _ := context.WithTimeout(context.TODO(), client.timeout)
		r = r.WithContext(ctx)
	}

	response, err = client.client.Do(r)

	return
}

func (client *HTTP2Client) Get(path string, headers http.Header) (b []byte, err error) {
	var response *http.Response
	if response, err = client.request("GET", path, nil, headers); err != nil {
		return
	}
	defer response.Body.Close()

	b, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	if response.StatusCode != http.StatusOK {
		err = errors.HTTPProblem.Clone().SetData("status", response.StatusCode).SetData("body", string(b))
		return
	}

	return
}

func (client *HTTP2Client) Post(path string, body []byte, headers http.Header) (b []byte, err error) {
	var bodyReader io.Reader

	if body != nil {
		bodyReader = bytes.NewBuffer(body)
	}

	var response *http.Response
	if response, err = client.request("POST", path, bodyReader, headers); err != nil {
		return
	}
	defer response.Body.Close()

	b, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	if response.StatusCode != http.StatusOK {
		err = errors.HTTPProblem.Clone().SetData("status", response.StatusCode).SetData("body", string(b))
		return
	}

	return
}
