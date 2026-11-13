package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	pb "protobuffer/pb"

	"github.com/rs/cors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

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
	w.Write(jsonData)
}

func main() {
	mux := http.NewServeMux()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	mux.HandleFunc("/fetch", fetchHandler)
	fmt.Println("Server live at localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", c.Handler(mux)))
}
