package RoundRobin

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

//     ___                    __  ___       __   _
//    / _ \___  __ _____  ___/ / / _ \___  / /  (_)__
//   / , _/ _ \/ // / _ \/ _  / / , _/ _ \/ _ \/ / _ \
//  /_/|_|\___/\_,_/_//_/\_,_/ /_/|_|\___/_.__/_/_//_/
//

type RrServer func()

type RoundRobinLoadbalancer struct {
	LbServer RrServer
}

func New() RoundRobinLoadbalancer {
	var mu sync.Mutex
	var idx int = 0

	// Loadbalancer Handler

	lbHandler := func(w http.ResponseWriter, r *http.Request) {
		maxLen := len(cfg.Backends)
		// Round Robin
		mu.Lock()
		currentBackend := cfg.Backends[idx%maxLen]
		targetURL, err := url.Parse(currentBackend.URL)
		if err != nil {
			logCh <- logEntry{time.Now(), logError, err.Error()}
			// log.Fatal(err.Error())
		}
		idx++
		mu.Unlock()
		reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
		reverseProxy.ServeHTTP(w, r)
		logCh <- logEntry{time.Now(), logInfo, "Requests Loaded to : http://" + targetURL.Host}

	}

	Server := func() {

		go logger()
		defer func() {
			close(logCh)
		}()

		data, err := ioutil.ReadFile("./config.json")
		if err != nil {
			logCh <- logEntry{time.Now(), logError, err.Error()}
			// log.Fatal(err.Error())
		}
		json.Unmarshal(data, &cfg)

		s := http.Server{
			Addr:    ":" + cfg.Proxy.Port,
			Handler: http.HandlerFunc(lbHandler),
		}
		logCh <- logEntry{time.Now(), logInfo, "Server up : http://localhost:" + cfg.Proxy.Port}
		if err = s.ListenAndServe(); err != nil {
			// log.Fatal(err.Error())
			logCh <- logEntry{time.Now(), logError, err.Error()}

		}
	}

	rr := RoundRobinLoadbalancer{LbServer: Server}
	return rr
}
