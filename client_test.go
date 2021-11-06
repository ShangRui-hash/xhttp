package xhttp

import (
	"context"
	"github.com/stretchr/testify/require"
	"github/jweny/xhttp/testutils"
	"golang.org/x/net/publicsuffix"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNewClient(t *testing.T) {
	options := NewDefaultClientOptions()
	client, err := NewClient(options, nil)
	require.Nil(t, err, "could not gen create HttpClient from options")
	require.NotEmpty(t, client, "could not gen http.Client from default")
}

func TestClient_Do(t *testing.T) {
	options := NewDefaultClientOptions()

	client, err := NewClient(options, nil)
	require.Nil(t, err, "could not gen create HttpClient from options")
	require.NotEmpty(t, client, "could not gen http.Client from default")

	ctx := context.Background()

	want := "success"

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header()
		w.WriteHeader(200)
		w.Write([]byte(want))
	}))

	hr, _ := http.NewRequest("GET", testServer.URL, nil)
	req := &Request{
		RawRequest: hr,
	}
	resp, err := client.Do(ctx, req)
	require.Nil(t, err, "could not do request with context")
	require.Equal(t, want, string(resp.Body), "could not get correct resp body")
}

func TestClient_Do_Redirect(t *testing.T) {
	ts := testutils.CreateRedirectServer(t)
	defer ts.Close()

	options := NewDefaultClientOptions()
	options.Headers = map[string]string{
		"user-agent": "aaa",
	}

	redirectClient, err := NewRedirectClient(options, nil)
	require.Nil(t, err, "could not new redirect http client")
	ctx := context.Background()
	hr, _ := http.NewRequest("GET", ts.URL + "/redirect-1", nil)
	req := &Request{
		RawRequest: hr,
	}
	resp, err := redirectClient.Do(ctx, req)
	require.Nil(t, err, "could not do request with redirect")
	body := resp.GetBody()
	require.Equal(t, "<a href=\"/redirect-11\">Temporary Redirect</a>.\n\n", string(body), "could not use redirect client")

	noRedirectClient, err := NewClient(options, nil)
	require.Nil(t, err, "could not new no redirect http client")
	resp1, err := noRedirectClient.Do(ctx, req)
	require.Nil(t, err, "could not do request with no redirect")
	body1 := resp1.GetBody()
	require.Equal(t, "<a href=\"/redirect-2\">Temporary Redirect</a>.\n\n", string(body1), "could not use redirect client")

}

func TestClient_Do_Cookie(t *testing.T) {
	ts := testutils.CreateRedirectServer(t)
	defer ts.Close()

	options := NewDefaultClientOptions()
	options.Headers = map[string]string{
		"user-agent": "aaa",
	}
	options.Cookies = map[string]string{
		"clientcookieid1": "id1",
		"clientcookieid2": "id2",
	}

	// 不跳转
	client, err := NewClient(options, nil)
	require.Nil(t, err, "could not new http client")
	ctx := context.Background()

	hr, _ := http.NewRequest("GET", ts.URL + "/redirect-1", nil)
	req := &Request{
		RawRequest: hr,
	}
	resp, err := client.Do(ctx, req)
	require.Nil(t, err, "could not do request with no redirect")
	require.Equal(t, "aaa", resp.Request.RawRequest.Header.Get("user-agent"))
	require.Equal(t, "clientcookieid1=id1; clientcookieid2=id2", resp.Request.RawRequest.Header.Get("cookie"))
}

func TestHeader_And_ResponseBodyLimit(t *testing.T) {
	ts := testutils.CreateGetServer(t)
	defer ts.Close()
	options := NewDefaultClientOptions()
	options.MaxRespBodySize = 100
	options.Cookies = map[string]string{
		"key1": "id1",
		"value1": "id2",
	}
	client, err := NewClient(options, nil)
	require.Nil(t, err, "could not new http client")
	ctx := context.Background()
	hr, _ := http.NewRequest("GET", ts.URL + "/", nil)
	req := &Request{
		RawRequest: hr,
	}
	req.SetHeader("user-agent", "aaa")
	req.EnableTrace()
	resp, err := client.Do(ctx, req)
	require.Nil(t, err, "could not do request with client")
	require.Equal(t, "aaa", resp.Request.GetHeaders().Get("user-agent"))
	require.Equal(t, "key1=id1; value1=id2", resp.Request.GetHeaders().Get("cookie"))
}

func TestAutoGzip(t *testing.T) {
	ts := testutils.CreateGenServer(t)
	defer ts.Close()

	options := NewDefaultClientOptions()
	client, err := NewClient(options, nil)
	require.Nil(t, err, "could not new http client")
	ctx := context.Background()

	testcases := []struct{ url, want string }{
		{ts.URL + "/gzip-test", "This is Gzip response testing"},
		{ts.URL + "/gzip-test-gziped-empty-body", ""},
		{ts.URL + "/gzip-test-no-gziped-body", ""},
	}
	for _, tc := range testcases {
		hr, _ := http.NewRequest("GET", tc.url, nil)
		req := &Request{
			RawRequest: hr,
		}
		resp, err := client.Do(ctx, req)
		require.Nil(t, err, "could not do request")
		body := resp.GetBody()
		require.Equal(t, tc.want, string(body), "could not auto gzip")
	}
}

func TestTransportCookie(t *testing.T) {
	ts := testutils.CreateGetServer(t)
	defer ts.Close()
	cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	options := NewDefaultClientOptions()
	//options.Debug = true
	ctx := context.Background()

	for i := 0; i <= 5; i++ {
		client, err := NewClient(options, cookieJar)
		require.Nil(t, err, "could not new http client")

		hr, _ := http.NewRequest("GET", ts.URL + "/transport-cookie", nil)
		req := &Request{
			RawRequest: hr,
		}
		req.SetHeader("user-agent", "aaa")
		req.EnableTrace()
		_, err = client.Do(ctx, req)
		require.Nil(t, err, "could not do request with client")
	}
	u, _ := url.Parse(ts.URL + "/transport-cookie")
	require.Equal(t, "success5", cookieJar.Cookies(u)[0].Value, "could not transport cookie to multi client")
}
