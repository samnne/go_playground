package repository

import (
	"fmt"
	"protobuffer/trips"
)

type TripRequest struct {
	date      string
	TripsData map[string]*trips.Trip
	count     int
}

var cache map[string]TripRequest = make(map[string]TripRequest)
var tripMap map[string]*trips.Trip

func GetTrips(routeID string, date string) map[string]*trips.Trip {
	key := routeID + date
	if tr, exists := cache[key]; exists {
		fmt.Printf("cache: %v\n", len(cache))
		return tr.TripsData
	}

	tripMap = trips.BuildTrips(routeID, date)
	return tripMap

}

func GetTripByID(id string, routeID string, date string) *trips.Trip {
	if trip, ok := tripMap[id]; ok {
		return trip
	}

	tripMap = GetTrips(routeID, date)
	return tripMap[id]
}

func AddToCache(routeID string, date string, tripMap map[string]*trips.Trip) {
	key := routeID + date
	temp, ok := cache[key]
	if ok && temp.count <= 100 {
		temp.count++

	} else {
		cache[key] = TripRequest{
			date:      date,
			TripsData: tripMap,
			count:     0,
		}
	}
}
