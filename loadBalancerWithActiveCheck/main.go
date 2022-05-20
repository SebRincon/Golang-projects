type Config struct {
    Proxy    Proxy     `json:"proxy"`
    Backends []Backend `json:"backends"`
}


type Proxy struct {
    Port string `json:"port"`
}

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

func loadBalancer(port string){
    maxLen := len(cfg.Backends)
    retry := 0

    // Round Robin overhead
    mu.Lock()
    currentBackend := cfg.Backends[idx%maxLen]
    if currentBackend.GetIsDead(){
        idx++
    }

    targetURL, err := url.Parse(cfg.Backends[idx%maxLen].URL)
    if err != nil{
        // LOG ERROR

    }

    idx++
    mu.Unlock()
    reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
    for (i := 0; i < retry; i++){
        
        reverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Requests, e error){
            // LOG ERROR
            currentBackend.SetDead(true)
            loadBalancer(w,r)

        }
    }
    reverseProxy.ServeHTTP(w,r)
    
}


var cfg Config

func readConfig(){
    

    data, err := ioutil.ReadFile("./config.json")    
    if err != nil {
        log.Fatal(err.Error())
    }
    json.Unmarshal(data, &cfg)
}
