package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	HttpClient *http.Client
	Headers    map[string]string
}

func NewClient(ProxyURL string, headers string) *Client {
	client := &http.Client{Timeout: 10 * time.Second}

	// set headers
	headerMap := func(headers string) map[string]string {
		headerMap := make(map[string]string)
		if headers != "" {
			headersPairs := strings.Split(headers, " -H ")
			for _, pair := range headersPairs {
				if pair != "" {
					parts := strings.SplitN(pair, ": ", 2)
					if len(parts) == 2 {
						headerMap[parts[0]] = parts[1]
					}
				}
			}
		}

		{
			// Add default headers that browsers usually set.
			headerMap["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:97.0) Gecko/20100101 Firefox/97.0"
			headerMap["Accept-Language"] = "en-US,en;q=0.5"
			headerMap["Connection"] = "keep-alive"
		}

		return headerMap
	}(headers)

	if ProxyURL != "" {
		proxy, err := url.Parse(ProxyURL)
		if err != nil {
			fmt.Printf("Error Parsing %s: %v", ProxyURL, err)
		} else {
			client.Transport = &http.Transport{Proxy: http.ProxyURL(proxy)}
		}
	}
	return &Client{HttpClient: client, Headers: headerMap}
}

func (c *Client) FetchURL(URL string) (body string, err error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return
	}

	for key, value := range c.Headers {
		req.Header.Add(key, value)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	body = string(bodyBytes)
	if err != nil {
		return
	}
	return
}
