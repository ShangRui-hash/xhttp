package http

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

/**
 * @Author: jweny
 * @Author: https://github.com/jweny
 * @Date: 2021/11/24 17:45
 * @Desc:
 */

func TestCreateGenServer(t *testing.T) {
	http.ListenAndServe(":12345", nil)
}

func TestUrlEncode(t *testing.T) {
	newURI, err := url.ParseRequestURI("/?${}")
	if err != nil {
		panic(err)
	}
	fmt.Println(newURI)
}
