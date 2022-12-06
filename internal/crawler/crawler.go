package crawler

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cr00z/goInnOgrnParser/internal/options"
	"github.com/jackdanger/collectlinks"
)

func CollectLinks(url string, body string) []string {
	correct := make([]string, 0)
	bodyReader := strings.NewReader(body)
	for _, link := range collectlinks.All(bodyReader) {
		link = strings.Trim(link, " ")
		link = strings.TrimPrefix(link, "https://"+url)
		link = strings.TrimPrefix(link, "http://"+url)
		if strings.HasPrefix(link, "http") {
			continue
		}
		if link == "" || link == "/" ||
			strings.HasPrefix(link, "tel:") ||
			strings.HasPrefix(link, "mailto:") ||
			strings.HasPrefix(link, "javascript:") {
			continue
		}
		if strings.HasSuffix(link, ".jpg") ||
			strings.HasSuffix(link, ".png") ||
			strings.HasSuffix(link, ".pdf") ||
			strings.HasSuffix(link, ".doc") ||
			strings.HasSuffix(link, ".docx") ||
			strings.HasSuffix(link, ".xls") ||
			strings.HasSuffix(link, ".xlsx") ||
			strings.HasSuffix(link, ".zip") ||
			strings.HasSuffix(link, ".rar") {
			continue
		}
		if link[0] != '/' {
			link = "/" + link
		}
		correct = append(correct, link)
	}
	return correct
}

func ParseNumbers(body string) map[string]int {
	candidates := make(map[string]int, 0)
	bodyLen := len(body)
	var pos, start int
	for pos < bodyLen-10 {
		// TODO: benchmark it
		if body[pos] < '0' || body[pos] > '9' {
			pos += 10
		} else {
			if pos < 10 {
				start = 0
			} else {
				start = pos
				for '0' <= body[start] && body[start] <= '9' {
					start--
				}
				start++
			}
			for pos < bodyLen && '0' <= body[pos] && body[pos] <= '9' {
				pos++
			}
			len := pos - start
			if len == 10 || len == 12 || len == 13 || len == 15 {
				candidates[body[start:pos]] = len
			}
		}
	}
	return candidates
}

func NewHttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext: (&net.Dialer{
				Timeout:   time.Second * time.Duration(options.HttpTimeout),
				KeepAlive: time.Second * time.Duration(options.HttpTimeout),
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          0,    // Default: 100
			MaxIdleConnsPerHost:   1000, // Default: 2
			IdleConnTimeout:       time.Second * time.Duration(options.HttpTimeout),
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: time.Second * time.Duration(options.HttpTimeout),
	}
}

func retrieveAttempt(httpClient *http.Client, url string, path string, proto string) (string, error) {
	req, err := http.NewRequest("GET", proto+url+path, nil)
	// TODO: Response errors: 4xx, 5xx
	if err != nil {
		return "", fmt.Errorf("transport error: %s", err)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("transport error: %s", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return "", errors.New("Response " + strconv.Itoa(res.StatusCode))
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read error: %s", err)
	}
	return string(body), nil
}

func Retrieve(httpClient *http.Client, url string, path string) (string, error) {
	resp, err := retrieveAttempt(httpClient, url, path, "https://")
	if err != nil {
		resp, err = retrieveAttempt(httpClient, url, path, "http://")
	}
	return resp, err
}

func HTTPErrorToString(err error) string {
	if strings.HasPrefix(err.Error(), "Response ") {
		return err.Error()
	}
	if strings.HasSuffix(err.Error(), "no such host") {
		return "No such host"
	}
	if strings.HasSuffix(err.Error(), "awaiting headers)") {
		return "Timeout"
	}
	if strings.HasSuffix(err.Error(), "connection refused") {
		return "Connection refused"
	}
	if strings.HasSuffix(err.Error(), "connection reset by peer") {
		return "Connection reset by peer"
	}
	if strings.HasSuffix(err.Error(), "network is unreachable") {
		return "Network is unreachable"
	}
	if strings.HasSuffix(err.Error(), " EOF") {
		return "Empty response"
	}
	if strings.HasSuffix(err.Error(), "tls: internal error") ||
		strings.HasSuffix(err.Error(), "tls: handshake failure") {
		return "SSL Protocol error"
	}
	if strings.HasSuffix(err.Error(), "stopped after 10 redirects") {
		return "Redirects"
	}
	if strings.HasSuffix(err.Error(), "no route to host") {
		return "No route to host"
	}
	return err.Error()
}
