package xhttp

import (
	"context"
	testhttp "github.com/jweny/xhttp/testutils/http"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestNewResponse(t *testing.T) {
	ts := testhttp.CreateGetServer(t)
	defer ts.Close()

	options := DefaultClientOptions()
	options.Cookies = map[string]string{
		"key1":   "id1",
		"value1": "id2",
	}
	ctx := context.Background()

	client, err := NewClient(options, nil)
	require.Nil(t, err, "could not new http client")

	hr, _ := http.NewRequest("GET", ts.URL+"/", nil)
	req := &Request{
		RawRequest: hr,
	}
	resp, err := client.Do(ctx, req)
	require.Nil(t, err)
	body := resp.GetBody()
	require.Nil(t, err)
	require.Equal(t, string(body), "TestGet: text response")
	dutation, err := resp.GetLatency()
	require.Nil(t, err)
	flag := false
	if dutation > 5 {
		flag = true
	}
	require.Equal(t, flag, true)
}
