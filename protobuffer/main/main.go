package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	pb "protobuffer/pb"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func main() {

	// "https://bct.tmix.se/gtfs-realtime/vehicleupdates.pb?operatorIds=48"
	http.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Printf("%v\n", r.RequestURI)
		req := &http.Request{
			Method: "GET",
			URL: &url.URL{
				Scheme:   "https",
				Host:     "bct.tmix.se",
				Path:     "/gtfs-realtime/vehicleupdates.pb",
				RawQuery: "operatorIds=48",
			},
			Header: make(http.Header),
		}
		// It is good practice to add standard headers if required by the API
		req.Header.Set("Accept", "application/x-protobuf")

		// 2. Send the request using the HTTP client
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic("Failed to make network request: " + err.Error())
		}
		defer resp.Body.Close()

		// 3. Verify the server returned a success status code
		if resp.StatusCode != http.StatusOK {
			// Tip: If the endpoint requires authentication, you will see a 401 or 403 here
			panic("Server returned bad status: " + resp.Status)
		}

		// 4. Read the raw binary body data
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			panic("Failed to read response body: " + err.Error())
		}

		// 5. Initialize your target struct and unmarshal the bytes
		msg := &pb.FeedMessage{}
		if err := proto.Unmarshal(bodyBytes, msg); err != nil {
			panic("Failed to unmarshal protobuf: " + err.Error())
		}

		jsonData, err := protojson.Marshal(msg)
		if err != nil {
			panic("Failed" + err.Error())
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	})

	http.ListenAndServe(":8080", nil)
}
