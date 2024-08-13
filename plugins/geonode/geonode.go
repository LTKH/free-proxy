package geonode

import (
    "fmt"
    //"crypto/tls"
    //"flag"
    //"io"
    //"log"
    //"net"
    "net/http"
    "time"
    "io/ioutil"
    "encoding/json"
    //"sync"
    "github.com/ltkh/free-proxy/internal/db"
    //"github.com/ltkh/free-proxy/internal/api"
    //"github.com/ltkh/free-proxy/internal/http"
)

type ProxyList struct {
    Data         []db.Proxy             `json:"data"`
}

/*
type Proxy struct {
    IP           string                 `json:"ip"`
    Port         string                 `json:"port"`
    Protocols    []string               `json:"protocols"`
    Latency      float64                `json:"latency"`
    ErrorCount   int64                  `json:"errorCount"`                
}
*/

var page = int(1);

//func init() {
//    log.Print("test geonode")
//}

func Gather() ([]db.Proxy, error) {
    client := &http.Client{
        Transport: &http.Transport{
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
            DisableCompression:  false,
        },
        Timeout: 5 * time.Second,
    }

    req, err := http.NewRequest("GET", fmt.Sprintf("https://proxylist.geonode.com/api/proxy-list?limit=500&page=%v&sort_by=lastChecked&sort_type=desc", page), nil)
    if err != nil {
        return nil, err
    }

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var proxyList ProxyList

    if err := json.Unmarshal(body, &proxyList); err != nil {
        return nil, err
    }

    //fmt.Printf("len - %v (%v)\n", len(proxyList.Data), page)

    if len(proxyList.Data) == 0 {
        page = 1;
    } else {
        page = page+1;
    }
    
    return proxyList.Data, nil
}