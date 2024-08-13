package http

import (
    "io"
    "log"
    "bytes"
    "net/http"
    "time"
    "io/ioutil"
    "fmt"
    "compress/gzip"
)

type HttpClient struct {
    client           *http.Client
    config           *HttpConfig
}

type HttpConfig struct {
    URL              string
    URLs             []string
    Headers          map[string]string
    ContentEncoding  string
    Username         string
    Password         string
}

type Response struct {
    Body             []byte
    StatusCode       int
    Header           http.Header
}

func New(config *HttpConfig) *HttpClient {
    client := &HttpClient{ 
        client: &http.Client{
            Transport: &http.Transport{
                MaxIdleConnsPerHost: 10,
                IdleConnTimeout:     90 * time.Second,
                DisableCompression:  false,
            },
            Timeout: 5 * time.Second,
        },
        config: config,
    }
    return client
}

func (h *HttpClient) NewRequest(method, path string, data []byte) (Response, error) {
    var resp Response
    var reader io.ReadCloser

    req, err := http.NewRequest(method, h.config.URL+path, bytes.NewReader(data))
    if err != nil {
        return resp, err
    }

    req.SetBasicAuth(h.config.Username, h.config.Password)
    req.Header.Set("Content-Type", "application/json")

    response, err := h.client.Do(req)
    if err != nil {
        return resp, err
    }
    resp.StatusCode = response.StatusCode
    resp.Header = response.Header
    
    // Check that the server actual sent compressed data
    switch response.Header.Get("Content-Encoding") {
        case "gzip":
            reader, err = gzip.NewReader(response.Body)
            if err != nil {
                return resp, err
            }
            defer reader.Close()
        default:
            reader = response.Body
    }

    body, err := ioutil.ReadAll(reader)
    if err != nil {
        return resp, err
    }
    resp.Body = body

    return resp, nil
}
