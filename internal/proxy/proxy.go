package proxy

import (
    //"os"
    //"fmt"
    "log"
    //"time"
    "sync"
    "net"
    "net/url"
    "net/http"
    "net/http/httputil"
    "io"
    "io/ioutil"
    "time"
    //"errors"
    //"regexp"
    "strings"
    "encoding/json"
    "compress/gzip"
    _ "github.com/mattn/go-sqlite3"
    //"github.com/ltkh/netmap/internal/config"
    "github.com/ltkh/free-proxy/internal/dbase"
    //"github.com/ltkh/free-proxy/internal/checker"
)

var (
    reverseProxy *httputil.ReverseProxy
    reverseProxyOnce sync.Once
    customTransport = http.DefaultTransport
)

//type proxy struct {
//    api              *api.Api
//}

type Api struct {
    db           *dbase.DB
}

type ReceivedData struct {
    Data         []dbase.Proxy        `json:"data"`
}

type Resp struct {
    Status       string               `json:"status"`
    Error        string               `json:"error,omitempty"`
    Warnings     []string             `json:"warnings,omitempty"`
    Data         interface{}          `json:"data"`
}

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

func encodeResp(resp *Resp) []byte {
    jsn, err := json.Marshal(resp)
    if err != nil {
        return encodeResp(&Resp{Status:"error", Error:err.Error(), Data:make([]int, 0)})
    }
    return jsn
}

func initReverseProxy() {
    proxyURL, err := url.Parse("http://127.0.0.1:8080")
    if err != nil {
        return
    }

    reverseProxy = &httputil.ReverseProxy{
        Director: func(r *http.Request) {
            //targetURL := "http://91.189.177.186:3128"
            target, err := url.Parse("https:"+r.URL.String())
            if err != nil {
                log.Printf("[error] unexpected error when parsing targetURL=%q: %s", r.URL.String(), err)
                return
            }
            target.Path = r.URL.Path
            target.RawQuery = r.URL.RawQuery
            r.URL = target
        },
        Transport: func() *http.Transport {
            tr := http.DefaultTransport.(*http.Transport).Clone()
            tr.DisableCompression = true
            tr.ForceAttemptHTTP2 = false
            tr.MaxIdleConnsPerHost = 100
            tr.Proxy = http.ProxyURL(proxyURL)
            if tr.MaxIdleConns != 0 && tr.MaxIdleConns < tr.MaxIdleConnsPerHost {
                tr.MaxIdleConns = tr.MaxIdleConnsPerHost
            }
            return tr
        }(),
        FlushInterval: time.Second,
        //ErrorLog:      logger.StdErrorLogger(),
        //ErrorLog:      log.New(new(bytes.Buffer), "", 0),
    }
}

func getReverseProxy() *httputil.ReverseProxy {
    reverseProxyOnce.Do(initReverseProxy)
    return reverseProxy
}

func NewProxy(rawUrl string) (*httputil.ReverseProxy, error) {
    log.Printf("rawUrl - %v", rawUrl)

    url, err := url.Parse(rawUrl)
    if err != nil {
        return nil, err
    }
    proxy := httputil.NewSingleHostReverseProxy(url)

    return proxy, nil
}

func New(db *dbase.DB) (*Api, error) {
    return &Api{ db: db }, nil
}

func (api *Api) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, " ", req.Method, " ", req.URL)

    req.URL.Scheme = "https"

	//if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
	//  	msg := "unsupported protocal scheme "+req.URL.Scheme
	//	http.Error(wr, msg, http.StatusBadRequest)
	//	log.Println(msg)
	//	return
	//}

	client := &http.Client{}

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go
	req.RequestURI = ""

	delHopHeaders(req.Header)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(wr, "Server Error", http.StatusInternalServerError)
		log.Fatal("ServeHTTP:", err)
	}
	defer resp.Body.Close()

	log.Println(req.RemoteAddr, " ", resp.Status)

	delHopHeaders(resp.Header)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
}

func (api *Api) ApiProxies(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var proxies []dbase.Proxy

    if r.Method == "POST" {
        var reader io.ReadCloser
        var err error

        // Check that the server actual sent compressed data
        switch r.Header.Get("Content-Encoding") {
            case "gzip":
                reader, err = gzip.NewReader(r.Body)
                if err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(400)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error(), Data:make([]int, 0)}))
                    return
                }
                defer reader.Close()
            default:
                reader = r.Body
        }
        defer r.Body.Close()

        body, err := ioutil.ReadAll(reader)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error(), Data:make([]int, 0)}))
            return
        }

        var data ReceivedData

        if err := json.Unmarshal(body, &data); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error(), Data:make([]int, 0)}))
            return
        }

        for _, proxy := range data.Data {
            if proxy.IP == "" {
                log.Print("[error] parameter missing IP")
                continue
            }
            if proxy.Port == "" {
                log.Print("[error] parameter missing port")
                continue
            }
            //test, _ := checker.ProxyTest("", fmt.Sprintf("%s://%s:%s", "socks4", proxy.IP, proxy.Port), 1)
            //if !test {
            //    log.Printf("[error] not connect - %v", fmt.Sprintf("%s://%s:%s", "socks4", proxy.IP, proxy.Port))
            //    continue
            //}
            proxies = append(proxies, proxy)
        }

        if err := api.db.SaveProxies(proxies); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error(), Data:make([]int, 0)}))
            return
        }
        
        w.WriteHeader(204)
        return
    }

    w.WriteHeader(405)
    w.Write(encodeResp(&Resp{Status:"error", Error:"method not allowed", Data:make([]int, 0)}))
}