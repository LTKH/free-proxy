package main

import (
    "fmt"
    //"crypto/tls"
    "flag"
    "os"
    "os/signal"
    //"io"
    "log"
    //"net"
    //"net/url"
    "net/http"
    //"net/http/httputil"
    "time"
    "sync"
    "syscall"
    "github.com/ltkh/free-proxy/plugins"
    "github.com/ltkh/free-proxy/internal/dbase"
    "github.com/ltkh/free-proxy/internal/proxy"
    "github.com/ltkh/free-proxy/internal/config"
    "github.com/ltkh/free-proxy/internal/checker"
)

/*
func handleTunneling(w http.ResponseWriter, r *http.Request) {
    log.Printf("handleTunneling - %v", r.RequestURI)

    dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }

    w.WriteHeader(http.StatusOK)
    hijacker, ok := w.(http.Hijacker)
    if !ok {
        http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
        return
    }

    client_conn, _, err := hijacker.Hijack()
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
    }

    go transfer(dest_conn, client_conn)
    go transfer(client_conn, dest_conn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
    defer destination.Close()
    defer source.Close()
    io.Copy(destination, source)
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
    log.Printf("handleHTTP - %v", r.URL.Path)

    resp, err := http.DefaultTransport.RoundTrip(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }

    defer resp.Body.Close()
    copyHeader(w.Header(), resp.Header)
    w.WriteHeader(resp.StatusCode)
    io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
    for k, vv := range src {
        for _, v := range vv {
            dst.Add(k, v)
        }
    }
}
*/

func checkProxies(debug bool, client *dbase.DB, proxies []dbase.Proxy) {
    var wg sync.WaitGroup

    targets := []string{
        //"https://www.youtube.com",
        "https://www.google.com",
    }

    for _, proxy := range proxies {
        go func(proxy dbase.Proxy, targets []string){
            wg.Add(1)
            defer wg.Done()

            count := int(0)
            laten := float64(0)

            for _, target := range targets {
                test, resp := checker.ProxyTest(fmt.Sprintf("%s://%s:%s", proxy.Protocol, proxy.IP, proxy.Port), target, 10)
                if test {
                    count = count + 1
                    laten = laten + resp
                }
            }

            if len(targets) > 0 { 
                percent := float64(count * 100 / len(targets))
                latency := laten / float64(len(targets))
                if debug == true {
                    log.Printf("[debug] %s://%s:%s - %v (%v)", proxy.Protocol, proxy.IP, proxy.Port, percent, latency)
                }
                
                proxy.Latency = latency
                client.UpdateProxy(percent, proxy)
            }

        }(proxy, targets)
    }
    wg.Wait()
}

func main() {
    pemPath := flag.String("pem", "server.pem", "path to pem file")
    keyPath := flag.String("key", "server.key", "path to key file")
    proto   := flag.String("proto", "http", "Proxy protocol (http or https)")
    debug   := flag.Bool("debug", false, "debug mode")
    flag.Parse()

    config.SetOsProxy(8888)

    if *proto != "http" && *proto != "https" {
        log.Fatal("[error] Protocol must be either http or https")
    }

    db, err := dbase.New("./config/proxy.db")
    if err != nil {
        log.Fatal(err)
    }
    db.CreateTables()

    plugins.New(db)

    pxy, err := proxy.New(db)
    if err != nil {
        log.Fatal(err)
    }

    //pxy, err := proxy.NewProxy("http://217.13.109.78:80")
    //if err != nil {
    //    log.Fatal(err)
    //}

    //handler := &proxy{ api: apiV1 }

    //http.HandleFunc("/api/v1/proxy/list", apiV1.ApiProxies)
    //http.HandleFunc("/", handler)
    //http.HandleFunc("/", handleTunneling)

    go func(){

        for {
            proxies, err := db.LoadProxies()
            if err != nil {
                log.Fatal(err)
            }

            list := []dbase.Proxy{}

            for k, proxy := range proxies {
                list = append(list, proxy)
                if len(list) >= 50 || k+1 == len(proxies) {
                    checkProxies(*debug, db, list)
                    list = nil
                }
            }

            time.Sleep(300 * time.Second)
        }
        
    }()

    /*
    server := &http.Server{
        Addr: "127.0.0.1:8888",
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.Method == http.MethodConnect {
                handleTunneling(w, r)
            } else {
                handleHTTP(w, r)
            }
        }),
        // Disable HTTP/2.
        TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
    }

    if *proto == "http" {
        log.Fatal(server.ListenAndServe())
    } else {
        log.Fatal(server.ListenAndServeTLS(*pemPath, *keyPath))
    }
    */

    // Program completion signal processing
    c := make(chan os.Signal, 2)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <- c
        config.UnsetOsProxy()
        log.Print("[info] server stopped")
        os.Exit(0)
    }()

    log.Print("[info] server started")

    if *proto == "https" {
        if err := http.ListenAndServeTLS("127.0.0.1:8888", *pemPath, *keyPath, pxy); err != nil {
            log.Fatalf("[error] %v", err)
        }
    } else {
        if err := http.ListenAndServe("127.0.0.1:8888", pxy); err != nil {
            log.Fatalf("[error] %v", err)
        }
    }
}
