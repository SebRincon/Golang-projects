package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
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

// Event Structure
// ----------------------->
type event struct {
	ID          string `json:"ID"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
}

type allEvents []event

var events = allEvents{
	{
		ID:          "1",
		Title:       "Default",
		Description: "......",
	},
}

// <-----------------------

// Wait Group
var logCh = make(chan logEntry, 50) // regular channel
var doneCh = make(chan struct{})    // signal only channel

// Callback Function
func homeLink(w http.ResponseWriter, r *http.Request) {
	logCh <- logEntry{time.Now(), logInfo, "API Has been Called"}
	fmt.Fprintf(w, "Welcome home!")

}

func main() {

	// Close logger at end of life
	go logger()
	defer func() {
		close(logCh)
	}()

	// Send data to logger
	logCh <- logEntry{time.Now(), logInfo, "App is Starting"}

	// Start server
	logCh <- logEntry{time.Now(), logInfo, "Server up : http://localhost:8080"}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeLink)
	router.HandleFunc("/event", createEvent).Methods("POST")
	router.HandleFunc("/events", getAllEvents).Methods("GET")
	router.HandleFunc("/events/{id}", getOneEvent).Methods("GET")
	router.HandleFunc("/events/{id}", updateEvent).Methods("PATCH")
	router.HandleFunc("/events/{id}", deleteEvent).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))

	// End of life
	logCh <- logEntry{time.Now(), logInfo, "App is Shutting Down"}
	time.Sleep(100 * time.Millisecond)
	// Send done signal
	doneCh <- struct{}{}
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

func createEvent(w http.ResponseWriter, r *http.Request) {
	var newEvent event
	reqBody, err := ioutil.ReadAll(r.Body)
	// If reading the body causes and error
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}
	// Parse json data to event type
	json.Unmarshal(reqBody, &newEvent)

	// Add Event to slice
	events = append(events, newEvent)

	// Send Back status and body
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newEvent)

	// log

	logCh <- logEntry{time.Now(), logInfo, "Event Created"}
}

func getOneEvent(w http.ResponseWriter, r *http.Request) {
	// Grab EventID
	eventID := mux.Vars(r)["id"]

	// Loop through slice and return event
	for _, singleEvent := range events {
		if singleEvent.ID == eventID {
			// log
			logString := fmt.Sprintf("Event # %v was queried", singleEvent.ID)
			logCh <- logEntry{time.Now(), logInfo, logString}
			// Send back event
			json.NewEncoder(w).Encode(singleEvent)

		}
	}

}

func getAllEvents(w http.ResponseWriter, r *http.Request) {
	// Log
	logCh <- logEntry{time.Now(), logInfo, "All Event have been queried"}
	// return entire slice
	json.NewEncoder(w).Encode(events)
}

func updateEvent(w http.ResponseWriter, r *http.Request) {
	// Get eventID
	eventID := mux.Vars(r)["id"]
	var updatedEvent event

	reqBody, err := ioutil.ReadAll(r.Body)
	// If and error is caused when fetching body
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}
	// Parse body to Event Type
	json.Unmarshal(reqBody, &updatedEvent)

	// Find Event by looping trough and update the properties
	for i, singleEvent := range events {
		if singleEvent.ID == eventID {
			singleEvent.Title = updatedEvent.Title
			singleEvent.Description = updatedEvent.Description
			events = append(events[:i], singleEvent)
			// Log
			logString := fmt.Sprintf("Event # %v was updated", singleEvent.ID)
			logCh <- logEntry{time.Now(), logInfo, logString}
			json.NewEncoder(w).Encode(singleEvent)
		}
	}
}

func deleteEvent(w http.ResponseWriter, r *http.Request) {
	// Fetch ID
	eventID := mux.Vars(r)["id"]

	// Loop until event is found, then delete the event
	for i, singleEvent := range events {
		if singleEvent.ID == eventID {
			events = append(events[:i], events[i+1:]...)
			logString := fmt.Sprintf("Event # %v was updated", singleEvent.ID)
			logCh <- logEntry{time.Now(), logInfo, logString}
		}
	}
}
