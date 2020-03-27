package main

import (
    "sync"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
)

type Request struct {
	Version string      `json:"version"`
	Op      string      `json:"op"`
	Content map[string]interface{} `json:"content"`
}

type Response struct {
	Reject       bool        `json:"reject"`
	RejectReason string      `json:"reject_reason"`
	Unchange     bool        `json:"unchange"`
	Content      map[string]interface{} `json:"content"`
}

type DomainInfo struct {
	passthrough bool
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

type APIServer struct {
	logger *log.Logger
    proxy  *ProxyServer
    domain string
    mutex sync.RWMutex
}

func (s APIServer) handler(w http.ResponseWriter, r *http.Request) {

        switch r.Method {
        case "POST":
                d := json.NewDecoder(r.Body)
                r := &Request{}
                o := &Response{}
                err := d.Decode(r)
                if err != nil {
                        http.Error(w, err.Error(), http.StatusInternalServerError)
                }

                if r.Op != "NewProxy" {
                    w.WriteHeader(http.StatusMethodNotAllowed)
                    fmt.Fprintf(w, "Not allowed.")
                    return
                }

                o.Reject = false
                o.Unchange = true

                if r.Content["proxy_type"] != "http" {
                    if r.Content["subdomain"] != "" {
                        var full_domain = s.domain + "." + r.Content["subdomain"].(string)
                        s.proxy.addFrontend(full_domain, false)
                    }

                    for _, domain := range r.Content["custom_domains"].([]string) {
                        s.proxy.addFrontend(domain, false)

                    }

                } else if  r.Content["proxy_type"] != "https" {
                    if r.Content["subdomain"] != "" {
                        var full_domain = s.domain + "." + r.Content["subdomain"].(string)
                        s.proxy.addFrontend(full_domain, true)
                    }

                    for _, domain := range r.Content["custom_domains"].([]string) {
                        s.proxy.addFrontend(domain, true)

                    }

                }

                js, err := json.Marshal(o)
                if err != nil {
                    http.Error(w, err.Error(), http.StatusInternalServerError)
                    return
                }
                w.Header().Set("Content-Type", "application/json")
                w.Write(js)

        default:
                w.WriteHeader(http.StatusMethodNotAllowed)
                fmt.Fprintf(w, "Not allowed.")
        }
}

func createAPIServer(logger *log.Logger, proxy *ProxyServer, port int, domain string) *APIServer {

    api := &APIServer{
        logger: logger,
        proxy: proxy,
        domain: domain,
    }

	http.HandleFunc("/", api.handler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

    return api

}

