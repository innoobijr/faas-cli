package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	gopath "path"
	"strings"
	"time"

	"github.com/innoobijr/faas-cli/version"
)

// Client an API client to perform all operations
type Client struct {
	httpClient *http.Client
	//ClientAuth a type implementing ClientAuth interface for client authentication
	ClientAuth ClientAuth
	//GatewayURL base URL of OpenFaaS gateway
	GatewayURL *url.URL
	//UserAgent user agent for the client
	UserAgent string
}

// ClientAuth an interface for client authentication.
// to add authentication to the client implement this interface
type ClientAuth interface {
	Set(req *http.Request) error
}

// NewClient initializes a new API client
func NewClient(auth ClientAuth, gatewayURL string, transport http.RoundTripper, timeout *time.Duration) (*Client, error) {
	gatewayURL = strings.TrimRight(gatewayURL, "/")
	baseURL, err := url.Parse(gatewayURL)
	if err != nil {
		return nil, fmt.Errorf("invalid gateway URL: %s", gatewayURL)
	}

	client := &http.Client{}
	if timeout != nil {
		client.Timeout = *timeout
	}

	if transport != nil {
		client.Transport = transport
	}

	return &Client{
		ClientAuth: auth,
		httpClient: client,
		GatewayURL: baseURL,
		UserAgent:  fmt.Sprintf("faas-cli/%s", version.BuildVersion()),
	}, nil
}

// newRequest create a new HTTP request with authentication
func (c *Client) newRequest(method, path string, query url.Values, body io.Reader) (*http.Request, error) {

	// deep copy gateway url and then add the supplied path  and args to the copy so that
	// we preserve the original gateway URL as much as possible
	endpoint, err := url.Parse(c.GatewayURL.String())
	if err != nil {
		return nil, err
	}

	endpoint.Path = gopath.Join(endpoint.Path, path)
	endpoint.RawQuery = query.Encode()

	bodyDebug := ""
	if os.Getenv("FAAS_DEBUG") == "1" {

		if body != nil {
			r := io.NopCloser(body)
			buf := new(strings.Builder)
			_, err := io.Copy(buf, r)
			if err != nil {
				return nil, err
			}
			bodyDebug = buf.String()
			body = io.NopCloser(strings.NewReader(buf.String()))
		}
	}

	req, err := http.NewRequest(method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	c.ClientAuth.Set(req)

	if os.Getenv("FAAS_DEBUG") == "1" {
		fmt.Printf("%s %s\n", req.Method, req.URL.String())
		for k, v := range req.Header {
			if k == "Authorization" {
				auth := "[REDACTED]"
				if len(v) == 0 {
					auth = "[NOT_SET]"
				} else {
					l, _, ok := strings.Cut(v[0], " ")
					if ok && (l == "Basic" || l == "Bearer") {
						auth = l + " REDACTED"
					}
				}
				fmt.Printf("%s: %s\n", k, auth)

			} else {
				fmt.Printf("%s: %s\n", k, v)
			}
		}

		if len(bodyDebug) > 0 {
			fmt.Printf("%s\n", bodyDebug)
		}
	}

	return req, err
}

// doRequest perform an HTTP request with context
func (c *Client) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)

	if val, ok := os.LookupEnv("OPENFAAS_DUMP_HTTP"); ok && val == "true" {
		dump, err := httputil.DumpRequest(req, true)
		if err != nil {
			return nil, err
		}
		fmt.Println(string(dump))
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}

	return res, err
}

func addQueryParams(u string, params map[string]string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return u, err
	}

	qs := parsedURL.Query()
	for key, value := range params {
		qs.Add(key, value)
	}
	parsedURL.RawQuery = qs.Encode()
	return parsedURL.String(), nil
}

// AddCheckRedirect add CheckRedirect to the client
func (c *Client) AddCheckRedirect(checkRedirect func(*http.Request, []*http.Request) error) {
	c.httpClient.CheckRedirect = checkRedirect
}
