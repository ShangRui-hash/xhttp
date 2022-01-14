package xhttp

import (
	"context"
	"errors"
	"fmt"
	"github.com/jweny/xhttp/xtls"
	"github.com/kataras/golog"
	"golang.org/x/net/http2"
	"golang.org/x/net/publicsuffix"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

type (
	// RequestMiddleware run before request
	RequestMiddleware func(*Request, *Client) error
	// ResponseMiddleware run after receive response
	ResponseMiddleware func(*Response, *Client) error
	// todo 未实现
	// errorHook after retry deal error
	errorHook func(*Request, error)
)

// Client struct
type Client struct {
	HTTPClient    *http.Client
	ClientOptions *ClientOptions
	Debug         bool        // if debug == true, start responseLogger middleware
	Error         interface{} // todo error handle exp
	// todo dns cache
	// Middleware
	defaultBeforeRequest []RequestMiddleware
	extraBeforeRequest   []RequestMiddleware
	afterResponse        []ResponseMiddleware
	errorHooks           []errorHook

	// handle
	closeConnection bool
}

// NewClient xhttp.Client
func NewClient(options *ClientOptions, jar *cookiejar.Jar) (*Client, error) {
	hc, err := createHttpClient(false, jar)
	if err != nil {
		return nil, err
	}

	client := createClient(options, hc)
	return client, nil
}

// NewRedirectClient xhttp.Client with Redirect
func NewRedirectClient(options *ClientOptions, jar *cookiejar.Jar) (*Client, error) {
	hc, err := createHttpClient(false, jar)
	if err != nil {
		return nil, err
	}

	client := createClient(options, hc)
	return client, nil
}

// NewDefaultClient xhttp.Client not follow redirect
func NewDefaultClient(jar *cookiejar.Jar) (*Client, error) {
	hc, err := createHttpClient(false, jar)
	if err != nil {
		return nil, err
	}

	client := createClient(GetHTTPOptions(), hc)
	return client, nil
}

// NewDefaultRedirectClient follow redirect
func NewDefaultRedirectClient(jar *cookiejar.Jar) (*Client, error) {
	hc, err := createHttpClient(true, jar)
	if err != nil {
		return nil, err
	}

	client := createClient(GetHTTPOptions(), hc)
	return client, nil
}

// NewWithHTTPClient with http client
func NewWithHTTPClient(options *ClientOptions, hc *http.Client) (*Client, error) {
	return createClient(options, hc), nil
}

// Do request
func (c *Client) Do(ctx context.Context, req *Request) (*Response, error) {
	var (
		resp                 *http.Response
		shouldRetry          bool
		err, doErr, retryErr error
	)
	if c == nil {
		return nil, errors.New("xhttp client not instantiated")
	}

	req.SetContext(ctx)
	req.attempt = 0

	err = c.ClientOptions.Limiter.Wait(req.GetContext())
	if err != nil {
		return nil, err
	}

	// user diy RequestMiddleware
	for _, f := range c.extraBeforeRequest {
		if err = f(req, c); err != nil {
			return nil, err
		}
	}

	// default diy RequestMiddleware
	for _, f := range c.defaultBeforeRequest {
		if err = f(req, c); err != nil {
			return nil, err
		}
	}
	// do request with retry
	for i := 0; ; i++ {
		req.attempt++

		req.setSendAt()
		resp, doErr = c.HTTPClient.Do(req.RawRequest)
		// need retry
		shouldRetry, retryErr = defaultRetryPolicy(req.GetContext(), resp, doErr)
		if !shouldRetry {
			break
		}
		remain := c.ClientOptions.FailRetries - i
		if remain <= 0 {
			break
		}
		// waitTime
		waitTime := defaultBackoff(defaultRetryWaitMin, defaultRetryWaitMax, i, resp)
		select {
		case <-time.After(waitTime):
		case <-req.GetContext().Done():
			return nil, req.GetContext().Err()
		}
	}

	if doErr == nil && retryErr == nil && !shouldRetry {
		// request success
		golog.Debugf("request: %s %s, response: status %d content-length %d", req.GetMethod(), req.GetUrl().String(), resp.StatusCode, resp.ContentLength)
		response := &Response{
			Request:     req,
			RawResponse: resp,
		}
		response.setReceivedAt()

		//ResponseMiddleware
		for _, f := range c.afterResponse {
			if err = f(response, c); err != nil {
				return nil, err
			}
		}
		return response, nil
	} else {
		// request fail
		finalErr := doErr
		if retryErr != nil {
			finalErr = retryErr
		}
		golog.Debugf("%s %s fail", req.GetMethod(), req.GetUrl().String())
		return nil, fmt.Errorf("giving up connect to %s %s after %d attempt(s): %v",
			req.RawRequest.Method, req.RawRequest.URL, req.attempt, finalErr)
	}
}

func (c *Client) BeforeRequest(fn RequestMiddleware) {
	c.extraBeforeRequest = append(c.extraBeforeRequest, fn)
}

func (c *Client) AfterResponse(fn ResponseMiddleware) {
	c.afterResponse = append(c.afterResponse, fn)
}

func (c *Client) SetCloseConnection(close bool) *Client {
	c.closeConnection = close
	return c
}

func createClient(options *ClientOptions, hc *http.Client) *Client {
	c := &Client{
		HTTPClient:    hc,
		ClientOptions: options,
		Debug:         options.Debug,
	}

	c.extraBeforeRequest = []RequestMiddleware{}
	c.defaultBeforeRequest = []RequestMiddleware{
		verifyRequestMethod,
		createHTTPRequest,
		readRequestBody,
	}
	c.afterResponse = []ResponseMiddleware{
		readResponseBody,
		verifyResponseBodyLength,
		responseLogger,
	}
	return c
}

func createHttpClient(followRedirects bool, jar *cookiejar.Jar) (*http.Client, error) {

	httpClientOptions := GetHTTPOptions()

	tlsClientConfig, err := xtls.NewTLSConfig(httpClientOptions.TlsOptions)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: time.Duration(httpClientOptions.DialTimeout) * time.Second,
		}).DialContext,
		MaxConnsPerHost:       httpClientOptions.MaxConnsPerHost,
		ResponseHeaderTimeout: time.Duration(httpClientOptions.ReadTimeout) * time.Second,
		IdleConnTimeout:       time.Duration(httpClientOptions.IdleConnTimeout) * time.Second,
		TLSHandshakeTimeout:   time.Duration(httpClientOptions.TLSHandshakeTimeout) * time.Second,
		MaxIdleConns:          httpClientOptions.MaxIdleConns,
		TLSClientConfig:       tlsClientConfig,
		DisableKeepAlives:     httpClientOptions.DisableKeepAlives,
	}
	if httpClientOptions.EnableHTTP2 {
		err := http2.ConfigureTransport(transport)
		if err != nil {
			return nil, err
		}
	}

	if httpClientOptions.Proxy != "" {
		proxy, err := url.Parse(httpClientOptions.Proxy)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxy)
	}
	// default cookiejar
	if jar == nil {
		cookieJar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			return nil, err
		}
		jar = cookieJar
	}

	return &http.Client{
		Jar:           jar,
		Transport:     transport,
		CheckRedirect: makeCheckRedirectFunc(followRedirects, httpClientOptions.MaxRedirect),
	}, nil
}

type checkRedirectFunc func(req *http.Request, via []*http.Request) error

func makeCheckRedirectFunc(followRedirects bool, maxRedirects int) checkRedirectFunc {
	return func(req *http.Request, via []*http.Request) error {
		if !followRedirects {
			return http.ErrUseLastResponse
		}
		if len(via) >= maxRedirects {
			return http.ErrUseLastResponse
		}
		return nil
	}
}
