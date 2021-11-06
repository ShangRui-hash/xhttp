package xhttp

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestNewRequest(t *testing.T) {
	testUrl := "http://127.0.0.1:54321/redirect"
	testMethod := MethodPost
	var testBody = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
	hr, _ := http.NewRequest(testMethod, testUrl, bytes.NewReader(testBody))
	req := &Request{
		RawRequest: hr,
	}
	require.Equal(t, testUrl, req.GetUrl().String(), "req.GetUrl id wrong")
	require.Equal(t, testMethod, req.GetMethod(), "req.GetMethod id wrong")
	requireBody, err := req.GetBody()
	require.Nil(t, err, "cannot use req.GetBody")
	require.Equal(t, testBody ,requireBody, "req.GetBody id wrong")
}
