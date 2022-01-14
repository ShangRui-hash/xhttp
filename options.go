package xhttp

import (
	"fmt"
	"github.com/jweny/xhttp/xtls"
	"golang.org/x/time/rate"
)

/**
 * @Author: jweny
 * @Author: https://github.com/jweny
 * @Date: 2021/12/2 16:36
 * @Desc:
 */

var HTTPOptions *ClientOptions

// ClientOptions http client options
type ClientOptions struct {
	Proxy string `json:"proxy" yaml:"proxy" #:"漏洞扫描时使用的代理, 如: http://127.0.0.1:8080. 如需设置多个代理, 请使用 proxy_rule 或自行创建上层代理"`
	//ProxyRule           []Rule       `json:"proxy_rule" yaml:"proxy_rule" #:"漏洞扫描使用多个代理的配置规则, 具体请参照文档"`
	DialTimeout         int  `json:"dial_timeout" yaml:"dial_timeout" #:"建立 tcp 连接的超时时间"`
	ReadTimeout         int  `json:"read_timeout" yaml:"read_timeout" #:"读取 http 响应的超时时间, 不可太小, 否则会影响到 sql 时间盲注的判断"`
	MaxConnsPerHost     int  `json:"max_conns_per_host" yaml:"max_conns_per_host" #:"同一 host 最大允许的连接数, 可以根据目标主机性能适当增大"`
	EnableHTTP2         bool `json:"enable_http2" yaml:"enable_http2" #:"是否启用 http2, 开启可以提升部分网站的速度, 但目前不稳定有崩溃的风险"`
	IdleConnTimeout     int  `json:"-" yaml:"-"`
	MaxIdleConns        int  `json:"-" yaml:"-"`
	TLSHandshakeTimeout int  `json:"-" yaml:"-"`

	FailRetries       int                 `json:"fail_retries" yaml:"fail_retries" #:"请求失败的重试次数, 0 则不重试"`
	MaxRedirect       int                 `json:"max_redirect" yaml:"max_redirect" #:"单个请求最大允许的跳转数"`
	MaxRespBodySize   int64               `json:"max_resp_body_size" yaml:"max_resp_body_size" #:"最大允许的响应大小, 默认 4M"`
	MaxQPS            int                 `json:"max_qps" yaml:"max_qps" #:"每秒最大请求数"`
	AllowMethods      []string            `json:"allow_methods" yaml:"allow_methods" #:"允许的请求方法"`
	Headers           map[string]string   `json:"headers" yaml:"headers" #:"自定义 headers"`
	Cookies           map[string]string   `json:"cookies" yaml:"cookies" #:"自定义 cookies, 参考 headers 格式， key: value"`
	TlsOptions        *xtls.ClientOptions `json:"tls" yaml:"tls" #:"tls 配置"`
	Debug             bool                `json:"http_debug" yaml:"http_debug" #:"是否启用 debug 模式, 开启 request trace"`
	DisableKeepAlives bool                `json:"disable_keep_alives" yaml:"disable_keep_alives" #:"是否禁用 keepalives"`
	Limiter           *rate.Limiter       `json:"-" yaml:"-"`
}

func (o *ClientOptions) SetLimiter() *ClientOptions {
	o.Limiter = rate.NewLimiter(rate.Limit(o.MaxQPS), 1)
	return o
}

func (o *ClientOptions) Verify() error {
	if o == nil || o.Limiter == nil {
		return fmt.Errorf("client options or Limiter cannot nil")
	}
	return nil
}

//type Server struct {
//	Options   string `json:"addr" yaml:"addr"`
//	Weight int    `json:"weight" yaml:"weight"`
//}
//
//type Rule struct {
//	Match   string   `json:"match" yaml:"match"`
//	Servers []Server `json:"servers" yaml:"servers"`
//}

const (
	// MethodGet HTTP method
	MethodGet = "GET"

	// MethodPost HTTP method
	MethodPost = "POST"

	// MethodPut HTTP method
	MethodPut = "PUT"

	// MethodDelete HTTP method
	MethodDelete = "DELETE"

	// MethodPatch HTTP method
	MethodPatch = "PATCH"

	// MethodHead HTTP method
	MethodHead = "HEAD"

	// MethodOptions HTTP method
	MethodOptions = "OPTIONS"

	// MethodConnect HTTP method
	MethodConnect = "CONNECT"

	// MethodTrace HTTP method
	MethodTrace = "TRACE"

	// MethodMove HTTP method
	MethodMove = "MOVE"

	// MethodPURGE MethodMove HTTP method
	MethodPURGE = "PURGE"

	//MethodPropFind = ""
)

func DefaultClientOptions() *ClientOptions {
	defaultHeaders := make(map[string]string)
	defaultHeaders["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; rv:78.0) Gecko/20100101 Firefox/78.0"

	return &ClientOptions{
		DialTimeout:         3,
		ReadTimeout:         10,
		IdleConnTimeout:     60,
		FailRetries:         0, // 默认改为0，否则如果配置文件指定了0，会不生效。 "nil value" 的问题
		MaxConnsPerHost:     50,
		MaxIdleConns:        50,
		TLSHandshakeTimeout: 5,
		MaxRedirect:         10,
		MaxRespBodySize:     2 << 20, // 4M
		MaxQPS:              500,
		Headers:             defaultHeaders,
		AllowMethods: []string{
			MethodHead,
			MethodGet,
			MethodPost,
			MethodPut,
			MethodPatch,
			MethodDelete,
			MethodOptions,
			MethodConnect,
			MethodTrace,
			MethodMove,
			MethodPURGE,
		},
		Cookies:           make(map[string]string),
		EnableHTTP2:       false,
		TlsOptions:        xtls.DefaultClientOptions(),
		Debug:             false,
		DisableKeepAlives: false,
		Limiter:           rate.NewLimiter(500, 1),
	}
}

func GetHTTPOptions() *ClientOptions {
	if HTTPOptions != nil {
		return HTTPOptions
	} else {
		return DefaultClientOptions()
	}
}
