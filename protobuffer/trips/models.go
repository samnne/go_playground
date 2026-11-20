package trips

import (
	"sync"
)

// Trip represents a GTFS trip from trips.txt.
// Columns: route_id,service_id,trip_id,trip_headsign,shape_id,block_id,direction_id
type Trip struct {
	RouteID      string `json:"route_id"`
	ServiceID    string `json:"service_id"`
	TripID       string `json:"trip_id"`
	TripHeadsign string `json:"trip_headsign"`
	ShapeID      string `json:"shape_id"`
	BlockID      string `json:"block_id"`
	DirectionID  int    `json:"direction_id"`
	Stops        []Stop `json:"stops"`
}

// Stop represents a stop on a trip, built from stop_times.txt joined with stops.txt.
type Stop struct {
	StopID       string  `json:"stop_id"`
	StopName     string  `json:"stop_name"`
	ArrivalTime  string  `json:"arrival_time"`
	StopSequence int     `json:"stop_sequence"`
	StopLat      float64 `json:"stop_lat"`
	StopLong     float64 `json:"stop_long"`
}

type stopDetails struct {
	name string
	lat  float64
	long float64
}

type Route struct {
	RouteColor string `json:"route_color"`
	RouteId    string `json:"route_id"`
	RouteName  string `json:"route_name"`
}

type gtfsStore struct {
	mu             sync.RWMutex
	loaded         bool
	activeServices map[string]map[string]struct{} // date -> service_id
	tripsByRoute   map[string][]*Trip             // route_id -> trips
	tripsByID      map[string]*Trip               // trip_id -> trip
	routes         map[string]*Route              // route_id -> route name
}

var store = &gtfsStore{
	activeServices: make(map[string]map[string]struct{}),
	tripsByRoute:   make(map[string][]*Trip),
	tripsByID:      make(map[string]*Trip),
	routes:         make(map[string]*Route),
}
