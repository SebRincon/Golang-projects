package PassiveCheck

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
	"net"
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

var cfg Config

//     ___               _            _______           __
//    / _ \___ ____ ___ (_)  _____   / ___/ /  ___ ____/ /__
//   / ___/ _ `(_-<(_-</ / |/ / -_) / /__/ _ \/ -_) __/  '_/
//  /_/   \_,_/___/___/_/|___/\__/  \___/_//_/\__/\__/_/\_\
//

// pingBackend checks if the backend is alive.
func isAlive(url *url.URL) bool {
	conn, err := net.DialTimeout("tcp", url.Host, time.Minute*1)
	if err != nil {

		logCh <- logEntry{time.Now(), logError, fmt.Sprintf("Unreachable to %v, error: %v", url.Host, err.Error())}
		return false
	}
	defer conn.Close()
	return true
}

// healthCheck is a function for healthcheck
func healthCheck() {
	t := time.NewTicker(time.Minute * 1)
	for {
		select {
		case <-t.C:
			for _, backend := range cfg.Backends {
				pingURL, err := url.Parse(backend.URL)
				if err != nil {
					logCh <- logEntry{time.Now(), logError, err.Error()}
				}
				isAlive := isAlive(pingURL)
				backend.SetDead(!isAlive)
				msg := "ok"
				if !isAlive {
					msg = "dead"
				}
				logCh <- logEntry{time.Now(), logInfo, fmt.Sprintf("%v checked %v by healthcheck", backend.URL, msg)}

			}
		}
	}

}

type PcServer func()

type PassiveCheckLoadbalancer struct {
	LbServer PcServer
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
		logCh <- logEntry{time.Now(), logError, err.Error()}
	}
	idx++
	mu.Unlock()
	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	reverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		// NOTE: It is better to implement retry.
		logCh <- logEntry{time.Now(), logError, fmt.Sprintf("%v is dead." ,targetURL)}
		currentBackend.SetDead(true)
		lbHandler(w, r)
	}
	reverseProxy.ServeHTTP(w, r)
	logCh <- logEntry{time.Now(), logInfo, "Requests Loaded to : http://" + targetURL.Host}
}

func New() PassiveCheckLoadbalancer {

	// Loadbalancer Handler
	// Serve serves a loadbalancer.
	Server := func() {
		go logger()
		defer func() {
			close(logCh)
		}()

		data, err := ioutil.ReadFile("./config.json")
		if err != nil {
			logCh <- logEntry{time.Now(), logError, err.Error()}
		}
		json.Unmarshal(data, &cfg)

		go healthCheck()

		s := http.Server{
			Addr:    ":" + cfg.Proxy.Port,
			Handler: http.HandlerFunc(lbHandler),
		}
		logCh <- logEntry{time.Now(), logInfo, "Server up : http://localhost:" + cfg.Proxy.Port}
		if err = s.ListenAndServe(); err != nil {
			logCh <- logEntry{time.Now(), logError, err.Error()}
		}
	}

	PcLb := PassiveCheckLoadbalancer{LbServer: Server}
	return PcLb

}
