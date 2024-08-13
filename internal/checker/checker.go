// Package checker provides a library for checking the availability of proxies.
package checker

import (
    "context"
    "net/http"
    URL "net/url"
    "time"
    //"log"
)

// ProxyTest checks the availability of a proxy by performing a test request.
//
// It sends an HTTP GET request to the specified target URL using the provided proxy.
// The function supports both HTTP and HTTPS proxies.
// It uses the specified timeout duration for the request.
//
// The function returns true if the proxy is available and the test request succeeds.
// It returns false if the proxy is unavailable or the test request fails.
//
// Example:
//
//    proxy := "http://1.1.1.1:8080"
//    targetURL := "http://example.com"
//    timeout := uint(5) // Timeout in seconds
//    result := ProxyTest(proxy, targetURL, timeout)
//    if result {
//        fmt.Println("Proxy is available")
//    } else {
//        fmt.Println("Proxy is not available")
//    }
func ProxyTest(proxy, urlTarget string, timeout uint) (bool, float64) {

    ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
    defer cancelCtx()

    client := &http.Client{Transport: &http.Transport{}}

    if proxy != "" {
        proxyURL, err := URL.Parse(proxy)
        if err != nil {
            return false, 0
        }

        client.Transport = &http.Transport{
            Proxy: http.ProxyURL(proxyURL),
        }
    }

    // Start timer
    start := time.Now()

    req, err := http.NewRequestWithContext(ctx, "GET", urlTarget, nil)
	if err != nil {
        return false, 0
    }
    req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64)")
    resp, err := client.Do(req)
    if err != nil {
        return false, 0
    }

    // Stop timer
    responseTime := time.Since(start).Seconds()

    if resp != nil {
        defer resp.Body.Close()
    }
    if resp.StatusCode >= 400 {
        return false, 0
    }

    return true, responseTime
}