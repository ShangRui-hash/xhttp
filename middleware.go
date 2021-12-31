package xhttp

import (
	"fmt"
	"github.com/thoas/go-funk"
	"io/ioutil"
	"net/http"
	"time"
)

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Request Middleware(s)
//_______________________________________________________________________

func verifyRequestMethod(req *Request, c *Client) error {
	// req.Method in AllowMethods
	currentMethod := req.RawRequest.Method
	if funk.Contains(c.ClientOptions.AllowMethods, currentMethod) == false {
		return fmt.Errorf(`http method %s not allowed`, currentMethod)
	}
	return nil
}

func readRequestBody(req *Request, c *Client) error {
	_, err := req.GetBody()
	if err != nil {
		return err
	}
	return nil
}

func createHTTPRequest(req *Request, c *Client) error {
	// enable trace
	if req.trace {
		req.clientTrace = &clientTrace{}
		req.ctx = req.clientTrace.createContext(req.GetContext())
	}
	// assign close connection option
	req.RawRequest.Close = c.closeConnection

	for key, value := range c.ClientOptions.Headers {
		req.RawRequest.Header.Set(key, value)
	}
	// fix header
	if req.RawRequest.Header.Get("Accept-Language") == "" {
		req.RawRequest.Header.Set("Accept-Language", "en")
	}
	if req.RawRequest.Header.Get("Accept") == "" {
		req.RawRequest.Header.Set("Accept", "*/*")
	}
	// add cookie
	if c.ClientOptions.Cookies != nil {
		for k, v := range c.ClientOptions.Cookies {
			req.RawRequest.AddCookie(&http.Cookie{
				Name:  k,
				Value: v,
			})
		}
	}
	// add ctx
	req.RawRequest = req.RawRequest.WithContext(req.GetContext())
	return nil
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Response Middleware(s)
//_______________________________________________________________________

func readResponseBody(resp *Response, c *Client) error {
	bodyBytes, err := ioutil.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return err
	}
	resp.Body = bodyBytes
	defer resp.RawResponse.Body.Close()
	return nil
}

func verifyResponseBodyLength(resp *Response, c *Client) error {
	// 检查响应长度
	length := int64(len(resp.Body))
	if length > c.ClientOptions.MaxRespBodySize {
		return fmt.Errorf("response Body longer %d than limit %d", length, c.ClientOptions.MaxRespBodySize)
	}
	return nil
}

func responseLogger(resp *Response, c *Client) error {
	if c.Debug {
		req := resp.Request
		reqString, err := req.GetRaw()
		if err != nil {
			return err
		}

		respString, err := resp.GetRaw()
		if err != nil {
			return err
		}

		latency, err := resp.GetLatency()
		if err != nil {
			return err
		}

		reqLog := "\n==============================================================================\n" +
			"--- REQUEST ---\n" +
			fmt.Sprintf("%s  %s  %s\n", req.GetMethod(), req.GetUrl().String(), req.RawRequest.Proto) +
			fmt.Sprintf("HOST   : %s\n", req.RawRequest.URL.Host) +
			fmt.Sprintf("RequestString:\n%s\n", reqString) +
			"------------------------------------------------------------------------------\n" +
			"--- RESPONSE ---\n" +
			fmt.Sprintf("STATUS       : %s\n", resp.RawResponse.Status) +
			fmt.Sprintf("PROTO        : %s\n", resp.RawResponse.Proto) +
			fmt.Sprintf("RECEIVED AT  : %v\n", resp.getReceivedAt().Format(time.RFC3339Nano)) +
			fmt.Sprintf("Attempt Num  : %d\n", req.attempt) +
			fmt.Sprintf("TIME DURATION: %v\n", latency) +
			fmt.Sprintf("HOST   : %s\n", req.RawRequest.URL.Host) +
			fmt.Sprintf("ResponseString:\n%s\n", respString) +
			"------------------------------------------------------------------------------\n"
		fmt.Println(reqLog)
	}
	return nil
}
