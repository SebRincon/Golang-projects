package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

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

// <-----------------------

// Wait Group
var logCh = make(chan logEntry, 50) // regular channel
var doneCh = make(chan struct{})    // signal only channel

func main() {

	// This is a way to pass in a port number to the program. Via Flags
	portFlag := flag.Int("port", 8081, "listening port")
	flag.Parse()
	port := fmt.Sprintf(":%d", *portFlag)

	// Close logger at end of life
	go logger()
	defer func() {
		close(logCh)
	}()

	loadBalancer(port)
}

func loadBalancer(port string) {
	var nextServerIndex int32 = 0
	var mu sync.Mutex

	// define origin server list to load balance the requests
	originServerList := []string{
		"http://localhost:8081",
		"http://localhost:8082",
	}

	loadBalancerHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// use mutex to prevent data race
		mu.Lock()

		// get next server to send a request to
		originServerURL, _ := url.Parse(originServerList[(nextServerIndex)%2])

		// increment next server value
		nextServerIndex++

		mu.Unlock()

		// use existing reverse proxy from httputil to route
		// a request to previously selected server url
		reverseProxy := httputil.NewSingleHostReverseProxy(originServerURL)

		reverseProxy.ServeHTTP(rw, req)
		logCh <- logEntry{time.Now(), logInfo, "Requst Loaded to : " + originServerURL.Host + ":" + originServerURL.Port()}

	})
	log.Fatal(http.ListenAndServe(port, loadBalancerHandler))
	logCh <- logEntry{time.Now(), logInfo, "Server up : http://localhost" + port}

}

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
