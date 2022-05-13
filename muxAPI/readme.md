# REST API Using Mux & Concurrent Logger

>This project uses [Mux]() to handle the routing based on the URL path and parameters, I also setup a mock database structure and mock endpoints to test the `GET`,`POST`,`DELETE`,& `PATCH` methods. In the background there is a logger that works by sending data to a channel and a concurrent fucntion prints the data out. 

### Mux Setup
---
This is how I have enabled flags when calling the the program from the terminal, by default the port will be 8081.
- Ex: `go run main.go -port=8080`
```
portFlag := flag.Int("port", 8081, "listening port")
flag.Parse()
port := fmt.Sprintf(":%d", *portFlag)
```
These are all the endpoints with params the API has and their callback functions.
```
router := mux.NewRouter().StrictSlash(true)
router.HandleFunc("/", homeLink)
router.HandleFunc("/event", createEvent).Methods("POST")
router.HandleFunc("/events", getAllEvents).Methods("GET")
router.HandleFunc("/events/{id}", getOneEvent).Methods("GET")
router.HandleFunc("/events/{id}", updateEvent).Methods("PATCH")
router.HandleFunc("/events/{id}", deleteEvent).Methods("DELETE")
log.Fatal(http.ListenAndServe(port, router))
```

### Logger Setup & Run
---
The logger setup is done setting up a message structure, this makes it easier to create log messages.
```
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
```
Two channels are then initialized, the `logCh` channel is used to pass in the message that will then be printed out. The `doneCh` channel is used to stop the concurrent log reader in order to not cause a panic at the end of the program. 
```
var logCh = make(chan logEntry, 50) // regular channel
var doneCh = make(chan struct{})    // signal only channel
```
This is how we would pass a message to the channel to be printed by the logger
```
logCh <- logEntry{time.Now(), logInfo, "App is Starting"}
logCh <- logEntry{time.Now(), logWarning, "This is a warning"}
logCh <- logEntry{time.Now(), logError, "There is an error"}
```
The logger function will loop through any channel messages and in this case either print out the message or if its a signal from `doneCh` it will end the function.
```
func logger() {
	for {
		select {
		case entry := <-logCh:
			fmt.Printf("%v : [%v] %v\n", entry.time.Format("2006-01-02"), entry.severity, entry.message)
		case <-doneCh:
			break
		}
	}
```
