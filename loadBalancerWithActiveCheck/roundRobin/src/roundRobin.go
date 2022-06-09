package roundRobin

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

//	   _____          ____
//	  / ___/__  ___  / _(_)__ _
//	 / /__/ _ \/ _ \/ _/ / _ `/
//	 \___/\___/_//_/_//_/\_, /
//          		    /___/

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

var cfg Config

//     ___                    __  ___       __   _
//    / _ \___  __ _____  ___/ / / _ \___  / /  (_)__
//   / , _/ _ \/ // / _ \/ _  / / , _/ _ \/ _ \/ / _ \
//  /_/|_|\___/\_,_/_//_/\_,_/ /_/|_|\___/_.__/_/_//_/
//

type RrServer func()

type RoundRobinLoadbalancer struct {
	rrServer RrServer
}

func New() RoundRobinLoadbalancer {
	var mu sync.Mutex
	var idx int = 0

	_lbHandler := func(w http.ResponseWriter, r *http.Request) {
		maxLen := len(cfg.Backends)
		// Round Robin
		mu.Lock()
		currentBackend := cfg.Backends[idx%maxLen]
		targetURL, err := url.Parse(currentBackend.URL)
		if err != nil {
			log.Fatal(err.Error())
		}
		idx++
		mu.Unlock()
		reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
		reverseProxy.ServeHTTP(w, r)
	}

	_rrServer := func() {
		data, err := ioutil.ReadFile("./config.json")
		if err != nil {
			log.Fatal(err.Error())
		}
		json.Unmarshal(data, &cfg)

		s := http.Server{
			Addr:    ":" + cfg.Proxy.Port,
			Handler: http.HandlerFunc(_lbHandler),
		}
		if err = s.ListenAndServe(); err != nil {
			log.Fatal(err.Error())
		}
	}

	rr := RoundRobinLoadbalancer{rrServer: _rrServer}
	return rr
}
