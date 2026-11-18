package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"protobuffer/middleware"
	"protobuffer/repository"
	"time"

	pb "protobuffer/pb"

	"github.com/gorilla/mux"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const feedURL = "https://bct.tmix.se/gtfs-realtime/vehicleupdates.pb?operatorIds=48"

var client = &http.Client{Timeout: 10 * time.Second}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func fetchHandler(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, feedURL, nil)
	if err != nil {
		http.Error(w, "bad upstream request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Accept", "application/x-protobuf")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("upstream request failed: %v", err)
		http.Error(w, "upstream request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("upstream returned %s", resp.Status)
		http.Error(w, "upstream returned "+resp.Status, http.StatusBadGateway)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("reading upstream body: %v", err)
		http.Error(w, "failed to read upstream", http.StatusBadGateway)
		return
	}

	msg := &pb.FeedMessage{}
	if err := proto.Unmarshal(body, msg); err != nil {
		log.Printf("decoding protobuf: %v", err)
		http.Error(w, "failed to decode feed", http.StatusBadGateway)
		return
	}

	jsonData, err := protojson.Marshal(msg)
	if err != nil {
		log.Printf("encoding json: %v", err)
		http.Error(w, "failed to encode feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	idx, err := w.Write(jsonData)
	fmt.Printf("idx: %v\n", idx)
	if err != nil {
		return
	}
}

func tripsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	routeID := query.Get("routeID")
	date := query.Get("date")
	fmt.Printf("date: %v\n", date)

	fmt.Printf("routeID: %v\n", routeID)
	if !query.Has("date") || !query.Has("routeID") {

		w.Write([]byte("Must include a routeID and date in url"))
		return
	}

	tripArray := repository.GetTrips(routeID, date)

	jsonData, err := json.Marshal(tripArray)
	if err != nil {
		log.Printf("encoding json: %v", err)
		http.Error(w, "failed to encode feed", http.StatusInternalServerError)
		return
	}
	repository.AddToCache(routeID, date, tripArray)

	w.Header().Set("Content-Type", "application/json")
	idx, err := w.Write(jsonData)
	fmt.Printf("idx: %v\n", idx)
	if err != nil {
		return
	}
}

func main() {

	router := mux.NewRouter()

	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.RebuildData)

	router.HandleFunc("/health", healthHandler)
	router.HandleFunc("/fetch", fetchHandler)
	router.HandleFunc("/trips", tripsHandler)
	router.HandleFunc("/trips/{id}", func(w http.ResponseWriter, r *http.Request) {
		vals := mux.Vars(r)
		fmt.Printf("id: %v\n", vals["id"])
		trip := repository.GetTripByID(vals["id"], "4-UVIC", "")
		jsonData, err := json.Marshal(trip)
		if err != nil {
			log.Printf("encoding json: %v", err)
			http.Error(w, "failed to encode feed", http.StatusInternalServerError)
			return
		}

		w.Write(jsonData)
	})
	router.Handle("/", http.FileServer(http.Dir("./client")))
	fmt.Println("Serving html on http://localhost:8080/")

	log.Fatal(http.ListenAndServe(":8080", router))
}
