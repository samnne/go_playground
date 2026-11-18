package tests

import (
	"testing"

	"protobuffer/repository"
	"protobuffer/trips"
)

func TestRepositoryGetTrips(t *testing.T) {
	tripMap := repository.GetTrips("4-VIC", "20261226")
	if len(tripMap) == 0 {
		t.Fatal("expected trips for 4-VIC on 20261226, got none")
	}
	if _, ok := tripMap["12337295:14530260:14827853"]; !ok {
		t.Fatal("expected known trip in repository results")
	}
}

func TestRepositoryGetTripByID(t *testing.T) {
	// Ensure trip map is populated via GetTrips (sets package-level tripMap on cache miss).
	repository.GetTrips("4-VIC", "20261226")

	trip := repository.GetTripByID("12337295:14530260:14827853", "4-VIC", "20261226")
	if trip == nil {
		t.Fatal("expected to find trip by ID")
	}
	if trip.TripID != "12337295:14530260:14827853" {
		t.Fatalf("expected matching TripID, got %q", trip.TripID)
	}
	if trip.RouteID != "4-VIC" {
		t.Fatalf("expected RouteID 4-VIC, got %q", trip.RouteID)
	}
}

func TestRepositoryAddToCacheAndGetTrips(t *testing.T) {
	routeID := "4-VIC"
	date := "20261226"

	tripMap := trips.BuildTrips(routeID, date)
	if len(tripMap) == 0 {
		t.Fatalf("expected trips for %s on %s to seed cache", routeID, date)
	}

	repository.AddToCache(routeID, date, tripMap)
	cached := repository.GetTrips(routeID, date)
	if len(cached) != len(tripMap) {
		t.Fatalf("expected cached trip count %d, got %d", len(tripMap), len(cached))
	}
	for id := range tripMap {
		if _, ok := cached[id]; !ok {
			t.Fatalf("expected cached trips to contain trip %s", id)
		}
	}
}
