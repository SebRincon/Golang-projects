# Load Balancer Explained

A load balancer is a kind of reverse proxy, which can be implemented with the httputil library from Golang's standard library

## Simple Server in Go
```
package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"
    "time"
)

func main() {
    portFlag := flag.Int("port", 8081, "listening port")
    flag.Parse()
    port := fmt.Sprintf(":%d", *portFlag)

    originServerHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
        fmt.Printf("[origin server] received request: %s\n", time.Now())
        _, _ = fmt.Fprintf(rw, "origin server response %s", port)
    })

    log.Fatal(http.ListenAndServe(port, originServerHandler))
}
```