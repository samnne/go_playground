package trips

import (
	"io"
	"strconv"
	"strings"
)

func readCalendarDates(r io.Reader) error {
	idx, reader, err := openCSVReader(r)
	if err != nil {
		return err
	}
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if rec[idx["exception_type"]] == "1" {
			d := rec[idx["date"]]
			sID := rec[idx["service_id"]]
			if store.activeServices[d] == nil {
				store.activeServices[d] = make(map[string]struct{})
			}
			store.activeServices[d][sID] = struct{}{}
		}
	}
	return nil
}

func readStopDetails(r io.Reader) (map[string]stopDetails, error) {
	idx, reader, err := openCSVReader(r)
	if err != nil {
		return nil, err
	}
	details := make(map[string]stopDetails)
	stopIDCol, hasID := idx["stop_id"]
	stopNameCol, hasName := idx["stop_name"]
	latCol, hasLat := idx["stop_lat"]
	lonCol, hasLon := idx["stop_lon"]
	if !hasLon {
		lonCol, hasLon = idx["stop_long"]
	}

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if !hasID || stopIDCol >= len(rec) {
			continue
		}
		sID := rec[stopIDCol]
		name := ""
		if hasName && stopNameCol < len(rec) {
			name = rec[stopNameCol]
		}
		var lat, long float64
		if hasLat && latCol < len(rec) {
			lat, _ = strconv.ParseFloat(rec[latCol], 64)
		}
		if hasLon && lonCol < len(rec) {
			long, _ = strconv.ParseFloat(rec[lonCol], 64)
		}
		details[sID] = stopDetails{
			name: name,
			lat:  lat,
			long: long,
		}
	}
	return details, nil
}

func readRoutes(r io.Reader) error {
	idx, reader, err := openCSVReader(r)
	if err != nil {
		return err
	}
	routeIDCol, ok := idx["route_id"]
	if !ok {
		return nil
	}
	longCol, hasLong := idx["route_long_name"]

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if routeIDCol >= len(rec) {
			continue
		}
		id := rec[routeIDCol]
		name := ""
		if hasLong && longCol < len(rec) {
			name = strings.TrimSpace(rec[longCol])
		}

		if id != "" {
			store.routes[id] = &Route{
				RouteColor: rec[idx["route_color"]],
				RouteId:    rec[idx["route_id"]],
				RouteName:  name,
			}
		}
	}
	return nil
}

func readTrips(r io.Reader) error {
	idx, reader, err := openCSVReader(r)
	if err != nil {
		return err
	}
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		dir := 0
		if v, ok := idx["direction_id"]; ok && v < len(rec) && rec[v] != "" {
			if n, err := strconv.Atoi(rec[v]); err == nil {
				dir = n
			}
		}

		t := &Trip{
			RouteID:      rec[idx["route_id"]],
			ServiceID:    rec[idx["service_id"]],
			TripID:       rec[idx["trip_id"]],
			TripHeadsign: rec[idx["trip_headsign"]],
			ShapeID:      rec[idx["shape_id"]],
			BlockID:      rec[idx["block_id"]],
			DirectionID:  dir,
			Stops:        make([]Stop, 0),
		}
		store.tripsByID[t.TripID] = t
		store.tripsByRoute[t.RouteID] = append(store.tripsByRoute[t.RouteID], t)
	}
	return nil
}

func readStopTimes(r io.Reader, stops map[string]stopDetails) error {
	idx, reader, err := openCSVReader(r)
	if err != nil {
		return err
	}
	tripCol := idx["trip_id"]
	stopCol := idx["stop_id"]
	arrCol := idx["arrival_time"]
	seqCol, hasSeq := idx["stop_sequence"]

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		t, ok := store.tripsByID[rec[tripCol]]
		if !ok {
			continue
		}
		seq := 0
		if hasSeq {
			if n, err := strconv.Atoi(rec[seqCol]); err == nil {
				seq = n
			}
		}
		sID := rec[stopCol]
		sd := stops[sID]
		t.Stops = append(t.Stops, Stop{
			StopID:       sID,
			StopName:     sd.name,
			ArrivalTime:  rec[arrCol],
			StopSequence: seq,
			StopLat:      sd.lat,
			StopLong:     sd.long,
		})
	}
	return nil
}
