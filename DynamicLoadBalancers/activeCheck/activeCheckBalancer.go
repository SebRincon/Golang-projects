package ActiveCheck


import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

//	   _____          ____
//	  / ___/__  ___  / _(_)__ _
//	 / /__/ _ \/ _ \/ _/ / _ `/
//	 \___/\___/_//_/_//_/\_, /
//          		    /___/

// Logging Structure
// ----------------------->
const (
	logInfo    = "INFO"
	logWarning = "WARNING"
	logError   = "ERROR"
)

type logEntry struct {
	time     time.Time
	severity string
	message  string
}

// Wait Group
var logCh = make(chan logEntry, 50) // regular channel
var doneCh = make(chan struct{})    // signal only channel



func logger() {
	for {
		select {
		case entry := <-logCh:
			fmt.Printf("%v : [%v] %v\n", entry.time.Format("2006-01-02"), entry.severity, entry.message)
		case <-doneCh:
			break
		}
	}
}

// <-----------------------


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

var cfg Config

//     ___      __  _            _______           __  
//    / _ |____/ /_(_)  _____   / ___/ /  ___ ____/ /__
//   / __ / __/ __/ / |/ / -_) / /__/ _ \/ -_) __/  '_/
//  /_/ |_\__/\__/_/|___/\__/  \___/_//_/\__/\__/_/\_\ 
//                                                     

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

type ACServer func()

type ActiveCheckLoadbalancer struct {
	LbServer ACServer
}




var mu sync.Mutex
var idx int = 0

	// lbHandler is a handler for loadbalancing
func lbHandler (w http.ResponseWriter, r *http.Request) {
		maxLen := len(cfg.Backends)
		// Round Robin
		mu.Lock()
		currentBackend := cfg.Backends[idx%maxLen]
		if currentBackend.GetIsDead() {
			idx++
		}
		targetURL, err := url.Parse(cfg.Backends[idx%maxLen].URL)
		if err != nil {
			logCh <- logEntry{time.Now(), logError, err.Error()}
		}
		idx++
		mu.Unlock()
		reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
		//! Active Check implemented with proxy Error Handler

		reverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			// NOTE: It is better to implement retry.
			logCh <- logEntry{time.Now(), logError, fmt.Sprintf("%v is dead." ,targetURL)}
			currentBackend.SetDead(true)
			lbHandler(w, r)
		}
		reverseProxy.ServeHTTP(w, r)
		logCh <- logEntry{time.Now(), logInfo, "Requests Loaded to : http://" + targetURL.Host}
	}

func New() ActiveCheckLoadbalancer {
	// Serve serves a loadbalancer.
	Server := func () {
		go logger()
		defer func() {
			close(logCh)
		}()

		data, err := ioutil.ReadFile("./config.json")
		if err != nil {
			logCh <- logEntry{time.Now(), logError, err.Error()}
		}
		json.Unmarshal(data, &cfg)

		s := http.Server{
			Addr:    ":" + cfg.Proxy.Port,
			Handler: http.HandlerFunc(lbHandler),
		}
		logCh <- logEntry{time.Now(), logInfo, "Server up : http://localhost:" + cfg.Proxy.Port}
		if err = s.ListenAndServe(); err != nil {
			logCh <- logEntry{time.Now(), logError, err.Error()}
		}
	}

	aclb := ActiveCheckLoadbalancer{LbServer: Server}
	return aclb

}


