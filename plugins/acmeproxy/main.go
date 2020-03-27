package main

import (
	"fmt"
	"log"
	"os"
    "golang.org/x/crypto/acme/autocert"
    "net/http"
    "net"
    "flag"
)

func getEnvString(key string, def string) string {
    val, ok := os.LookupEnv(key)
    if !ok {
        return def
    } else {
        return val
    }
}

func redirectHttps(w http.ResponseWriter, r *http.Request){
    host, _, _ := net.SplitHostPort(r.Host)
    u := r.URL
    u.Host = host
    u.Scheme="https"
    log.Println(u.String())
    http.Redirect(w,r,u.String(), http.StatusMovedPermanently)
}

func main() {

    var apiPort = flag.Int("api", 9000, "API port on localhost")
    var mainDomain = flag.String("domain", "", "Main domain for subdomains")

    flag.Parse()

    logger := log.New(os.Stdout, "acmeproxy ", log.LstdFlags|log.Lshortfile)

	m := &autocert.Manager{
		Cache:      autocert.DirCache("certs"),
		Prompt:     autocert.AcceptTOS,
	}

	s := &ProxyServer{
		Logger:  logger,
        Manager: m,
	}

    createAPIServer(logger, s, *apiPort, *mainDomain)

	err := s.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}

    http.ListenAndServe(":80", http.HandlerFunc(redirectHttps))

}

