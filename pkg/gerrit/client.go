package gerrit

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type Config struct {
	Server    string
	Username  string
	Password  string
	Insecure  bool
	Debug     bool
	NoAuthPrefix bool // skip /a/ prefix (for nginx-authed instances)
}

type ErrUnexpectedResponse struct {
	Status     string
	StatusCode int
	Body       string
}

func (e *ErrUnexpectedResponse) Error() string {
	return fmt.Sprintf("unexpected response: %s - %s", e.Status, e.Body)
}

type ClientFunc func(*Client)

type Client struct {
	server       string
	username     string
	password     string
	insecure     bool
	debug        bool
	noAuthPrefix bool
	timeout      time.Duration
	transport    http.RoundTripper
}

func NewClient(cfg Config, opts ...ClientFunc) *Client {
	c := &Client{
		server:       strings.TrimSuffix(cfg.Server, "/"),
		username:     cfg.Username,
		password:     cfg.Password,
		insecure:     cfg.Insecure,
		debug:        cfg.Debug,
		noAuthPrefix: cfg.NoAuthPrefix,
		timeout:      15 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	c.transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.insecure,
		},
		DialContext: (&net.Dialer{Timeout: c.timeout}).DialContext,
	}
	return c
}

func WithTimeout(d time.Duration) ClientFunc {
	return func(c *Client) { c.timeout = d }
}

func WithInsecureTLS(b bool) ClientFunc {
	return func(c *Client) { c.insecure = b }
}

func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *Client) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

func (c *Client) Put(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.do(ctx, http.MethodPut, path, body)
}

func (c *Client) Delete(ctx context.Context, path string) error {
	_, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	prefix := "/a/"
	if c.noAuthPrefix {
		prefix = "/"
	}
	url := fmt.Sprintf("%s%s%s", c.server, prefix, strings.TrimPrefix(path, "/"))

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Transport: c.transport}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if c.debug {
		reqDump, _ := httputil.DumpRequest(req, true)
		respDump, _ := httputil.DumpResponse(resp, false)
		fmt.Printf("\n--- REQUEST ---\n%s\n--- RESPONSE ---\n%s\n", reqDump, respDump)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &ErrUnexpectedResponse{
			Status:     resp.Status,
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}
	}

	respBody = bytes.TrimPrefix(respBody, []byte(")]}'"))
	respBody = bytes.TrimLeft(respBody, "\n")
	return respBody, nil
}
