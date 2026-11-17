package trips

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
)

// Trip represents a GTFS trip from trips.txt.
// Columns: route_id,service_id,trip_id,trip_headsign,shape_id,block_id,direction_id
type Trip struct {
	RouteID      string
	ServiceID    string
	TripID       string
	TripHeadsign string
	ShapeID      string
	BlockID      string
	DirectionID  int
	Stops        []Stop
}

// Stop represents a stop on a trip, built from stop_times.txt joined with stops.txt.
type Stop struct {
	StopID       string
	StopName     string
	ArrivalTime  string
	StopSequence int
}

func openCSVReader(file io.Reader, err error) (map[string]int, *csv.Reader) {
	if err != nil {
		return nil, nil
	}

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, nil
	}

	return indexByName(header), reader
}

// GetActiveServiceIDs reads calendar_dates.txt and returns the set of service_ids
// active on the given date (YYYYMMDD) with exception_type == "1".
func GetActiveServiceIDs(date string) map[string]struct{} {
	file, err := os.Open("data/calendar_dates.txt")
	idx, reader := openCSVReader(file, err)
	active := make(map[string]struct{})

	defer file.Close()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if record[idx["date"]] == date && record[idx["exception_type"]] == "1" {
			active[record[idx["service_id"]]] = struct{}{}
		}
	}

	return active
}

// BuildStops reads stop_times.txt for tripID and joins stop names from stops.txt.
func attachStops(trips []Trip, wanted map[string]int) {
	if len(wanted) == 0 {
		return
	}
	stopNames := loadStopNames() // once

	file, err := os.Open("data/stop_times.txt")
	idx, reader := openCSVReader(file, err)
	defer file.Close()

	tripCol := idx["trip_id"]
	stopCol := idx["stop_id"]
	arrCol := idx["arrival_time"]
	seqCol, hasSeq := idx["stop_sequence"]

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		i, ok := wanted[record[tripCol]]
		if !ok {
			continue
		}

		seq := 0
		if hasSeq {
			if n, err := strconv.Atoi(record[seqCol]); err == nil {
				seq = n
			}
		}

		stopID := record[stopCol]
		trips[i].Stops = append(trips[i].Stops, Stop{
			StopID:       stopID,
			StopName:     stopNames[stopID],
			ArrivalTime:  record[arrCol],
			StopSequence: seq,
		})
	}

	for i := range trips {
		s := trips[i].Stops
		sort.Slice(s, func(a, b int) bool { return s[a].StopSequence < s[b].StopSequence })
	}
}

func loadStopNames() map[string]string {
	file, err := os.Open("data/stops.txt")
	idx, reader := openCSVReader(file, err)
	defer file.Close()

	names := make(map[string]string)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		names[record[idx["stop_id"]]] = record[idx["stop_name"]]
	}
	return names
}

// BuildTrips reads trips.txt and returns trips for routeID that are active on date.
// If date is empty, all trips for routeID are returned.
// main entry point
func BuildTrips(routeID, date string) []Trip {
	file, err := os.Open("data/trips.txt")
	idx, reader := openCSVReader(file, err)
	defer file.Close()

	var active map[string]struct{}
	if date != "" {
		active = GetActiveServiceIDs(date)
	}

	var tripsOut []Trip
	wanted := make(map[string]int) // tripID -> index in tripsOut

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if routeID != "" && record[idx["route_id"]] != routeID {
			continue
		}
		serviceID := record[idx["service_id"]]
		if active != nil {
			if _, ok := active[serviceID]; !ok {
				continue
			}
		}

		dir := 0
		if v, ok := idx["direction_id"]; ok && v < len(record) && record[v] != "" {
			if n, err := strconv.Atoi(record[v]); err == nil {
				dir = n
			}
		}

		tripID := record[idx["trip_id"]]
		wanted[tripID] = len(tripsOut)
		tripsOut = append(tripsOut, Trip{
			RouteID:      record[idx["route_id"]],
			ServiceID:    serviceID,
			TripID:       tripID,
			TripHeadsign: record[idx["trip_headsign"]],
			ShapeID:      record[idx["shape_id"]],
			BlockID:      record[idx["block_id"]],
			DirectionID:  dir,
		})
	}

	attachStops(tripsOut, wanted)
	return tripsOut
}

func indexByName(header []string) map[string]int {
	idx := make(map[string]int, len(header))
	for i, name := range header {
		// Strip UTF-8 BOM if present on first column.
		if len(name) > 0 && name[0] == '\xef' {
			name = name[3:]
		}
		idx[name] = i
	}
	return idx
}
