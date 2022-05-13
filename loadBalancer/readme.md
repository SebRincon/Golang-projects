# Load Balancer Explained
>A load balancer is a kind of [reverse proxy](), which allows us to dynamically route HTTP requests to specific endpoints. This package only containts the loadbalancer so refer to [muxAPI]() or [Simple API in Go]() to create several API instaces if needed for testing.
                                                  
                                                                                            
                                                                                            
                                ┏━━━━━━━━━━━━━━━┓                                           
                                ┃               ┃     ┌───────────────┐   ┌───────────────┐ 
                             ┌─▷┃    Client     ┃     │   Server 1    │░  │   Server 2    │░
                             │  ┃               ┃     │localhost:8081 │░  │localhost:8082 │░
                             │  ┗━━━━━━━━━━━━━━━┛     │               │░  │               │░
                        HTTP Request                  └───────────────┘░  └───────────────┘░
                             │  ┌───────────────┐      ░░░░░░░▲░░░░░░░░░   ░░░░░░░▲░░░░░░░░░
                             │  │ Load Balancer │             │                   │         
                             └─▶│localhost:8080 │◁────────────┼───────────────────┤         
                                │               │             │                   │         
                                └───────────────┘             ▼                   ▼         
                                                      ┌───────────────┐   ┌───────────────┐ 
                                                      │   Server 3    │░  │   Server 4    │░
                                                      │localhost:8083 │░  │localhost:8084 │░
                                                      │               │░  │               │░
                                                      └───────────────┘░  └───────────────┘░
                                                       ░░░░░░░░░░░░░░░░░   ░░░░░░░░░░░░░░░░░
                                                                                                                                                   
                                                                                            
                                                                                            
 

### Config 
In `main.go` > `main` function I have enabled a port flag to specify the loadbalancer's port, Ex: `go run main.go -port=8080` by default the port is 8080.
```
portFlag := flag.Int("port", 8080, "listening port")
flag.Parse()
port := fmt.Sprintf(":%d", *portFlag)

loadBalancer(port)
```

In `main.go` > `loadBalancer` function I have specified what are the endpoints that requests should be forwarded to. This is where endpoints can be added or changed.

```
originServerList := []string{
	"http://localhost:8081",
	"http://localhost:8082",
}

```

### Simple API in Go
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
	// Extract Port Flag, w/ 8081 as default & format to string
	portFlag := flag.Int("port", 8081, "listening port")
	flag.Parse()
	port := fmt.Sprintf(":%d", *portFlag)
		
	  
	// Request Handler
	originServerHandler := http.HandlerFunc(func(rw http.ResponseWriter,
	req*http.Request) {
		
		fmt.Printf("[origin server] received request: %s\n", time.Now())
				_, _ = fmt.Fprintf(rw, "origin server response %s", port)
	
	})
	
	  
	// Listen and Serve	
	log.Fatal(http.ListenAndServe(port, originServerHandler))

}

```
