package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	pb "protobuffer/pb"
	"protobuffer/trips"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type TripRequest struct {
	date      string
	TripsData *[]trips.Trip
}

var cache map[string]TripRequest = make(map[string]TripRequest)

const feedURL = "https://bct.tmix.se/gtfs-realtime/vehicleupdates.pb?operatorIds=48"

var client = &http.Client{Timeout: 10 * time.Second}

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

		jsonData, err := json.Marshal("")
		if err != nil {
			return
		}

		w.Write(jsonData)
		return
	}

	if tr, exists := cache[routeID]; exists && tr.date == date {

		jsonData, err := json.Marshal(tr.TripsData)
		if err != nil {
			log.Printf("encoding json: %v", err)
			http.Error(w, "failed to encode feed", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonData)
		if err != nil {
			return
		}
		return
	}

	tripArray := trips.BuildTrips(routeID, date)

	jsonData, err := json.Marshal(tripArray)
	if err != nil {
		log.Printf("encoding json: %v", err)
		http.Error(w, "failed to encode feed", http.StatusInternalServerError)
		return
	}
	cache[routeID] = TripRequest{
		date:      date,
		TripsData: &tripArray,
	}
	w.Header().Set("Content-Type", "application/json")
	idx, err := w.Write(jsonData)
	fmt.Printf("idx: %v\n", idx)
	if err != nil {
		return
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/fetch", fetchHandler)
	mux.HandleFunc("/trips", tripsHandler)
	mux.Handle("/", http.FileServer(http.Dir("./client")))
	fmt.Println("Serving html on http://localhost:8080/")

	log.Fatal(http.ListenAndServe(":8080", mux))
}
