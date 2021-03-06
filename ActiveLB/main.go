package main

import (
	"encoding/json"

	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"log"

)

// Config is a configuration.
type Config struct {
    Proxy    Proxy     `json:"proxy"`
    Backends []Backend `json:"backends"`
}

// Proxy is a reverse proxy, and means load balancer.
type Proxy struct {
    Port string `json:"port"`
}


// Backend is servers which load balancer is transferred.
type Backend struct {
    URL    string `json:"url"`
    IsDead bool
    mu     sync.RWMutex
}

// SetDead updates the value of IsDead in Backend.
func (backend *Backend) SetDead(b bool) {
    backend.mu.Lock()
    backend.IsDead = b
    backend.mu.Unlock()
}

// GetIsDead returns the value of IsDead in Backend.
func (backend *Backend) GetIsDead() bool {
    backend.mu.RLock()
    isAlive := backend.IsDead
    backend.mu.RUnlock()
    return isAlive
}

var mu sync.Mutex
var idx int = 0

// lbHandler is a handler for loadbalancing
func lbHandler(w http.ResponseWriter, r *http.Request) {
    maxLen := len(cfg.Backends)
    // Round Robin
    mu.Lock()
    currentBackend := cfg.Backends[idx%maxLen]
    if currentBackend.GetIsDead() {
        idx++
    }
    targetURL, err := url.Parse(cfg.Backends[idx%maxLen].URL)
    if err != nil {
        log.Fatal(err.Error())
    }
    idx++
    mu.Unlock()
    reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
    reverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
        // NOTE: It is better to implement retry.
        log.Printf("%v is dead.", targetURL)
        currentBackend.SetDead(true)
        lbHandler(w, r)
    }
    reverseProxy.ServeHTTP(w, r)
}

var cfg Config

// Serve serves a loadbalancer.
func Serve() {
    data, err := ioutil.ReadFile("./config.json")
    if err != nil {
        log.Fatal(err.Error())
    }
    json.Unmarshal(data, &cfg)

    s := http.Server{
        Addr:    ":" + cfg.Proxy.Port,
        Handler: http.HandlerFunc(lbHandler),
    }
    if err = s.ListenAndServe(); err != nil {
        log.Fatal(err.Error())
    }
}


func main(){
	Serve()
}
