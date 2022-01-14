package http

import (
	"compress/gzip"
	"fmt"
	"github.com/kataras/golog"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func getTestDataPath() string {
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, ".testdata")
}

func createTestServer(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}

func CreateGetServer(t *testing.T) *httptest.Server {
	var attempt int32
	var sequence int32
	var lastRequest time.Time
	ts := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		//t.Logf("Method: %v", r.Method)
		//t.Logf("Path: %v", r.URL.Path)

		if r.Method == "GET" {
			switch r.URL.Path {
			case "/index.action/struts/utils.js":
				value := r.Header.Get("If-Modified-Since")
				if value != "" {
					golog.Infof(value)
					reg := regexp.MustCompile(`(?m):\/\/(.*?)\/`)
					matches := reg.FindAllSubmatch([]byte(value), -1)
					if len(matches) > 0 {
						domain := string(matches[0][1])
						golog.Infof(domain)
						cmd := exec.Command("ping", domain)
						err := cmd.Run()
						if err != nil {
							golog.Fatalf("cmd.Run() failed with %s\n", err)
						}
						_, _ = w.Write([]byte("success"))
					} else {
						_, _ = w.Write([]byte("error"))
					}
				} else {
					_, _ = w.Write([]byte("no header"))
				}
			case "/XMII/Catalog":
				_, _ = w.Write([]byte("root:*:0:0:System Administrator:/var/root:/bin/sh"))
			case "/":
				_, _ = w.Write([]byte("TestGet: text response"))
			case "/no-content":
				_, _ = w.Write([]byte(""))
			case "/json":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"TestGet": "JSON response"}`))
			case "/json-invalid":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte("TestGet: Invalid JSON"))
			case "/long-text":
				_, _ = w.Write([]byte("TestGet: text response with size > 30"))
			case "/long-json":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"TestGet": "JSON response with size > 30"}`))
			case "/mypage":
				w.WriteHeader(http.StatusBadRequest)
			case "/mypage2":
				_, _ = w.Write([]byte("TestGet: text response from mypage2"))
			case "/set-retrycount-test":
				attp := atomic.AddInt32(&attempt, 1)
				if attp <= 4 {
					time.Sleep(time.Second * 5)
				}
				_, _ = w.Write([]byte("TestClientRetry page"))
			case "/set-retrywaittime-test":
				// Returns time.Duration since last request here
				// or 0 for the very first request
				if atomic.LoadInt32(&attempt) == 0 {
					lastRequest = time.Now()
					_, _ = fmt.Fprint(w, "0")
				} else {
					now := time.Now()
					sinceLastRequest := now.Sub(lastRequest)
					lastRequest = now
					_, _ = fmt.Fprintf(w, "%d", uint64(sinceLastRequest))
				}
				atomic.AddInt32(&attempt, 1)

			case "/set-retry-error-recover":
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				if atomic.LoadInt32(&attempt) == 0 {
					w.WriteHeader(http.StatusTooManyRequests)
					_, _ = w.Write([]byte(`{ "message": "too many" }`))
				} else {
					_, _ = w.Write([]byte(`{ "message": "hello" }`))
				}
				atomic.AddInt32(&attempt, 1)
			case "/set-timeout-test-with-sequence":
				seq := atomic.AddInt32(&sequence, 1)
				time.Sleep(time.Second * 2)
				_, _ = fmt.Fprintf(w, "%d", seq)
			case "/set-timeout-test":
				time.Sleep(time.Second * 6)
				_, _ = w.Write([]byte("TestClientTimeout page"))
			case "/my-image.png":
				fileBytes, _ := ioutil.ReadFile(filepath.Join(getTestDataPath(), "test-img.png"))
				w.Header().Set("Content-Type", "image/png")
				w.Header().Set("Content-Length", strconv.Itoa(len(fileBytes)))
				_, _ = w.Write(fileBytes)
			case "/get-method-payload-test":
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Errorf("Error: could not read get body: %s", err.Error())
				}
				_, _ = w.Write(body)
			case "/host-header":
				_, _ = w.Write([]byte(r.Host))
			case "/transport-cookie":
				//fmt.Printf("第 %d 次请求的Cookie： ", attempt)
				//fmt.Println(r.Cookies())
				//设置cookie
				tNow := time.Now()
				cookie := &http.Cookie{
					Name:    "totem",
					Value:   "success" + strconv.Itoa(int(attempt)),
					Expires: tNow.AddDate(1, 0, int(attempt)),
				}
				http.SetCookie(w, cookie)
				//返回信息
				_, _ = w.Write([]byte("your cookie has been received"))
				atomic.AddInt32(&attempt, 1)
			}

			switch {
			case strings.HasPrefix(r.URL.Path, "/v1/users/sample@sample.com/100002"):
				if strings.HasSuffix(r.URL.Path, "details") {
					_, _ = w.Write([]byte("TestGetPathParams: text response: " + r.URL.String()))
				} else {
					_, _ = w.Write([]byte("TestPathParamURLInput: text response: " + r.URL.String()))
				}
			}

		}
	})

	return ts
}

func CreateGenServer(t *testing.T) *httptest.Server {
	ts := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		//t.Logf("Method: %v", r.Method)
		//t.Logf("Path: %v", r.URL.Path)

		if r.Method == "GET" {
			if r.URL.Path == "/json-no-set" {
				// Set empty header value for testing, since Go server sets to
				// text/plain; charset=utf-8
				w.Header().Set("Content-Type", "")
				_, _ = w.Write([]byte(`{"response":"json response no content type set"}`))
			} else if r.URL.Path == "/gzip-test" {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.Header().Set("Content-Encoding", "gzip")
				zw := gzip.NewWriter(w)
				_, _ = zw.Write([]byte("This is Gzip response testing"))
				zw.Close()
			} else if r.URL.Path == "/gzip-test-gziped-empty-body" {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.Header().Set("Content-Encoding", "gzip")
				zw := gzip.NewWriter(w)
				// write gziped empty body
				_, _ = zw.Write([]byte(""))
				zw.Close()
			} else if r.URL.Path == "/gzip-test-no-gziped-body" {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.Header().Set("Content-Encoding", "gzip")
				// don't write body
			}

			return
		}

		if r.Method == "PUT" {
			if r.URL.Path == "/plaintext" {
				_, _ = w.Write([]byte("TestPut: plain text response"))
			} else if r.URL.Path == "/json" {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				_, _ = w.Write([]byte(`{"response":"json response"}`))
			} else if r.URL.Path == "/xml" {
				w.Header().Set("Content-Type", "application/xml")
				_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Response>XML response</Response>`))
			}
			return
		}

		if r.Method == "OPTIONS" && r.URL.Path == "/options" {
			w.Header().Set("Access-Control-Allow-Origin", "localhost")
			w.Header().Set("Access-Control-Allow-Methods", "PUT, PATCH")
			w.Header().Set("Access-Control-Expose-Headers", "x-go-resty-id")
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "PATCH" && r.URL.Path == "/patch" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "REPORT" && r.URL.Path == "/report" {
			body, _ := ioutil.ReadAll(r.Body)
			if len(body) == 0 {
				w.WriteHeader(http.StatusOK)
			}
			return
		}
	})

	return ts
}

func CreateRedirectServer(t *testing.T) *httptest.Server {
	ts := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		//t.Logf("Method: %v", r.Method)
		//t.Logf("Path: %v", r.URL.Path)

		if r.Method == "GET" {
			if strings.HasPrefix(r.URL.Path, "/redirect-host-check-") {
				cntStr := strings.SplitAfter(r.URL.Path, "-")[3]
				cnt, _ := strconv.Atoi(cntStr)

				if cnt != 7 { // Testing hard stop via logical
					if cnt >= 5 {
						http.Redirect(w, r, "http://httpbin.org/get", http.StatusTemporaryRedirect)
					} else {
						http.Redirect(w, r, fmt.Sprintf("/redirect-host-check-%d", cnt+1), http.StatusTemporaryRedirect)
					}
				}
			} else if strings.HasPrefix(r.URL.Path, "/redirect-") {
				cntStr := strings.SplitAfter(r.URL.Path, "-")[1]
				cnt, _ := strconv.Atoi(cntStr)

				http.Redirect(w, r, fmt.Sprintf("/redirect-%d", cnt+1), http.StatusTemporaryRedirect)
			}
		}
	})

	return ts
}
