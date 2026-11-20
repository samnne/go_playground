package repository

import (
	"fmt"
	"sync"

	"protobuffer/trips"
)

type TripRequest struct {
	date      string
	TripsData map[string]*trips.Trip
	count     int
}

var (
	cacheMu sync.RWMutex
	cache   = make(map[string]TripRequest)
	tripMap map[string]*trips.Trip
)

// LoadGTFS loads and indexes GTFS data into memory at startup.
// The path can be a data directory or a zip file (streamed from archive entries).
func LoadGTFS(path string) error {
	return trips.LoadGTFS(path)
}

// GetRouteName returns the route name for a route ID.
func GetRoute(routeID string) *trips.Route {
	return trips.GetRoute(routeID)
}

func GetTrips(routeID string, date string) map[string]*trips.Trip {
	key := routeID + date

	cacheMu.RLock()
	tr, exists := cache[key]
	cacheMu.RUnlock()
	if exists {
		fmt.Printf("cache: %v\n", len(cache))
		return tr.TripsData
	}

	result := trips.BuildTrips(routeID, date)

	cacheMu.Lock()
	tripMap = result
	cacheMu.Unlock()

	return result
}

func GetTripByID(id string, routeID string, date string) *trips.Trip {
	if trip := trips.GetTripByID(id); trip != nil {
		// If a date filter is provided, ensure the trip is active for that date+route context.
		if date == "" || routeID == "" || (trip.RouteID == routeID && tripInDate(trip, date)) {
			return trip
		}
	}

	cacheMu.RLock()
	if trip, ok := tripMap[id]; ok {
		cacheMu.RUnlock()
		return trip
	}
	cacheMu.RUnlock()

	tm := GetTrips(routeID, date)
	return tm[id]
}

func AddToCache(routeID string, date string, tripsData map[string]*trips.Trip) {
	key := routeID + date

	cacheMu.Lock()
	defer cacheMu.Unlock()

	temp, ok := cache[key]
	if ok && temp.count <= 100 {
		temp.count++
		cache[key] = temp
	} else {
		cache[key] = TripRequest{
			date:      date,
			TripsData: tripsData,
			count:     0,
		}
	}
}

func tripInDate(trip *trips.Trip, date string) bool {
	active := trips.GetActiveServiceIDs(date)
	_, ok := active[trip.ServiceID]
	return ok
}
