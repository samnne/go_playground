package trips

import (
	"io"
	"log"
	"os"
	"strconv"
)

// GetActiveServiceIDs returns the set of service_ids active on the given date
// (YYYYMMDD) with exception_type == "1".
func GetActiveServiceIDs(date string) map[string]struct{} {
	store.mu.RLock()
	if store.loaded {
		defer store.mu.RUnlock()
		out := make(map[string]struct{})
		for sid := range store.activeServices[date] {
			out[sid] = struct{}{}
		}
		return out
	}
	store.mu.RUnlock()

	file, err := os.Open("data/calendar_dates.txt")
	if err != nil {
		return map[string]struct{}{}
	}
	idx, reader := openCSVReaderLegacy(file, err)
	active := make(map[string]struct{})
	defer file.Close()
	if reader == nil {
		return active
	}

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

// BuildTrips returns trips for routeID that are active on date.
// If date is empty, all trips for routeID are returned.
func BuildTrips(routeID, date string) map[string]*Trip {
	store.mu.RLock()
	if store.loaded {
		defer store.mu.RUnlock()
		candidates := store.tripsByRoute[routeID]
		activeSet := store.activeServices[date]
		out := make(map[string]*Trip)
		for _, t := range candidates {
			if date != "" {
				if _, ok := activeSet[t.ServiceID]; !ok {
					continue
				}
			}
			out[t.TripID] = t
		}
		return out
	}
	store.mu.RUnlock()

	// Fallback to legacy on-disk behavior if not preloaded.
	file, err := os.Open("data/trips.txt")
	if err != nil {
		return map[string]*Trip{}
	}
	idx, reader := openCSVReaderLegacy(file, err)
	defer file.Close()
	if reader == nil {
		return map[string]*Trip{}
	}

	var active map[string]struct{}
	if date != "" {
		active = GetActiveServiceIDs(date)
	}

	tripsOut := make(map[string]*Trip)
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
		tripsOut[tripID] = &Trip{
			RouteID:      record[idx["route_id"]],
			ServiceID:    serviceID,
			TripID:       tripID,
			TripHeadsign: record[idx["trip_headsign"]],
			ShapeID:      record[idx["shape_id"]],
			BlockID:      record[idx["block_id"]],
			DirectionID:  dir,
		}
	}

	attachStopsLegacy(tripsOut)
	return tripsOut
}

// GetTripByID returns a trip by ID from the in-memory store if loaded.
func GetTripByID(id string) *Trip {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if !store.loaded {
		return nil
	}
	return store.tripsByID[id]
}

// GetRouteName returns the route name for a route ID from the in-memory store.
// If the store is not loaded, it attempts to read from data/routes.txt.
func GetRoute(routeID string) *Route {
	store.mu.RLock()
	if store.loaded {
		route := store.routes[routeID]
		store.mu.RUnlock()
		return route
	}
	route := store.routes[routeID]
	store.mu.RUnlock()
	if route != nil {
		return route
	}

	f, err := os.Open("data/routes.txt")
	if err != nil {
		return nil
	}
	defer f.Close()

	readRoutes(f)
	return store.routes[routeID]
}
