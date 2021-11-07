package xhttp

import (
	"context"
	"github.com/jweny/xhttp/testutils"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestNewResponse(t *testing.T) {
	ts := testutils.CreateGetServer(t)
	defer ts.Close()

	options := NewDefaultClientOptions()
	ctx := context.Background()

	client, err := NewClient(options, nil)
	require.Nil(t, err, "could not new http client")

	hr, _ := http.NewRequest("GET", ts.URL + "/set-retrycount-test", nil)
	req := &Request{
		RawRequest: hr,
	}
	resp, err := client.Do(ctx, req)
	require.Nil(t, err)
	body := resp.GetBody()
	require.Nil(t, err)
	require.Equal(t, string(body), "TestClientRetry page")
	duration, err := resp.GetLatency()
	require.Nil(t, err)
	flag := false
	if duration > 5 *time.Second {
		flag = true
	}
	require.Equal(t, flag, true, "response latency is wrong")
}
