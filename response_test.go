package xhttp

import (
	"context"
	"github.com/stretchr/testify/require"
	"github/jweny/xhttp/testutils"
	"net/http"
	"testing"
)

func TestNewResponse(t *testing.T) {
	ts := testutils.CreateGetServer(t)
	defer ts.Close()

	options := NewDefaultClientOptions()
	options.Debug = true
	options.Cookies = map[string]string{
		"key1": "id1",
		"value1": "id2",
	}
	ctx := context.Background()

	client, err := NewClient(options, nil)
	require.Nil(t, err, "could not new http client")

	hr, _ := http.NewRequest("GET", ts.URL + "/set-retrycount-test", nil)
	req := &Request{
		RawRequest: hr,
	}
	req.SetHeader("user-agent", "aaa")

	resp, err := client.Do(ctx, req)
	require.Nil(t, err)
	body := resp.GetBody()
	require.Nil(t, err)
	require.Equal(t, string(body), "TestClientRetry page")
	duration, err := resp.GetLatency()
	require.Nil(t, err)
	flag := false
	if duration > 5 {
		flag = true
	}
	require.Equal(t, flag, true)
}
