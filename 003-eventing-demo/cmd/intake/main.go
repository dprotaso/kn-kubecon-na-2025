package main

import (
	"cmp"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"kndemo"
	"log"
	"net/http"
	"os"
	"sync"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	ceclient "github.com/cloudevents/sdk-go/v2/client"
	"github.com/google/uuid"
)

//go:embed static/*
//go:embed index.html
var content embed.FS

var (
	clients = make(map[chan SSEEvent]bool)
	mu      sync.Mutex
)

type SSEEvent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func main() {
	h := New().Handle
	http.HandleFunc("/complain", h)
	http.Handle("/", http.FileServerFS(content))
	http.HandleFunc("/cloudevent", ceHandler)
	http.HandleFunc("/events", handleSSE)

	http.ListenAndServe(":"+cmp.Or(os.Getenv("PORT"), "8080"), nil)
}

// MyFunction is the function provided by this library.
// This structure name can be changed.
type MyFunction struct {
	ce ceclient.Client
}

// New constructs an instance of your function.  It is called each time a new
// instance of the function service is created.  This function must be named
// "New", accept no arguments, and return a structure which exports at least
// a Handle method (and optionally any of the additional methods described
// in the comments below).
func New() *MyFunction {
	if os.Getenv("K_SINK") == "" {
		log.Fatalf("K_SINK env var must be defined")
	}

	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	return &MyFunction{ce: c}
}

// Handle a request using your function instance.
func (f *MyFunction) Handle(w http.ResponseWriter, r *http.Request) {
	var payload kndemo.Payload
	payload.Content = r.FormValue("complaint")

	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())

	// This ID serves as the unique identifier for the message
	event.SetSubject(event.Context.GetID())
	event.SetSource("intake")
	event.SetType("intake.new")

	// var data *kndemo.Payload
	// dec := json.NewDecoder(r.Body)
	//
	// if err := dec.Decode(&data); err != nil {
	// 	log.Println("failed to read body ", err)
	// 	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	// 	return
	// }

	event.SetData(cloudevents.ApplicationJSON, payload)
	bytes, _ := json.Marshal(event)

	log.Println("Sending ", string(bytes))
	// Set a target.
	ctx := cloudevents.ContextWithTarget(context.Background(), os.Getenv("K_SINK"))

	// Send that Event.
	if result := f.ce.Send(ctx, event); cloudevents.IsUndelivered(result) {
		log.Printf("failed to send, %v", result)
	} else {
		log.Printf("sent: %v", event)
		log.Printf("result: %v", result)
	}

	w.WriteHeader(http.StatusAccepted)
}

func handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	msgChan := make(chan SSEEvent)
	mu.Lock()
	clients[msgChan] = true
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(clients, msgChan)
		mu.Unlock()
		close(msgChan)
	}()

	for {
		select {
		case msg := <-msgChan:
			log.Printf("writing message %#v", msg)
			data, _ := json.Marshal(msg)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func ceHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("got cloud event")
	event, err := cloudevents.NewEventFromHTTPRequest(r)
	if err != nil {
		log.Printf("failed to parse CloudEvent from request: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	var payload *kndemo.Payload
	event.DataAs(&payload)

	broadcast(SSEEvent{Type: "apology", Text: payload.Apology})
}

func broadcast(event SSEEvent) {
	log.Printf("broadcasting event %#v", event)
	mu.Lock()
	defer mu.Unlock()
	for ch := range clients {
		select {
		case ch <- event:
		default:
			delete(clients, ch)
			close(ch)
		}
	}
}
