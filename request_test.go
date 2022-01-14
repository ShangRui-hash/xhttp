package xhttp

import (
	"bytes"
	"context"
	"fmt"
	testhttp "github.com/jweny/xhttp/testutils/http"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"
)

func TestNewRequest(t *testing.T) {
	ts := testhttp.CreateGenServer(t)
	defer ts.Close()

	restUrl := ts.URL + "/json-no-set"

	testMethod := MethodPost
	var testBody = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
	hr, _ := http.NewRequest(testMethod, restUrl, bytes.NewReader(testBody))
	req := &Request{
		RawRequest: hr,
	}
	require.Equal(t, restUrl, req.GetUrl().String(), "req.GetUrl id wrong")
	require.Equal(t, testMethod, req.GetMethod(), "req.GetMethod id wrong")
	requireBody, err := req.GetBody()
	require.Nil(t, err, "cannot use req.GetBody")
	require.Equal(t, testBody, requireBody, "req.GetBody id wrong")

	raw, err := req.GetRaw()
	require.Nil(t, err, "cannot use req.GetRaw")
	fmt.Println("1", string(raw))

	opt := DefaultClientOptions()
	opt.MaxRespBodySize = 100
	opt.Cookies = map[string]string{
		"key1":   "id1",
		"value1": "id2",
	}
	client, err := NewClient(opt, nil)
	require.Nil(t, err, "could not new http client")
	ctx := context.Background()
	resp, err := client.Do(ctx, req)
	require.Nil(t, err, "could not new http client")

	raw2, err := resp.Request.GetRaw()
	fmt.Println("2", string(raw2))
	//require.Contains(t, string(requireRaw), string(testBody), "raw is wrong")
}

func TestSetPostParam(t *testing.T) {
	hr, _ := http.NewRequest("POST", "http://192.168.123.30:12345", nil)
	req := Request{RawRequest: hr}

	var params = map[string]string{
		"p1": "111",
		"p2": "222",
	}
	var paramList []string
	for k, v := range params {
		paramList = append(paramList, fmt.Sprintf("%s=%s", k, v))
	}
	p := strings.Join(paramList, "&")
	req.SetBody([]byte(p))
	raw, _ := req.GetRaw()
	fmt.Println(string(raw))
}
