package tests

import (
	"os"
	"testing"

	"protobuffer/trips"
)

func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestGetActiveServiceIDs(t *testing.T) {
	active := trips.GetActiveServiceIDs("20261226")
	if active == nil {
		t.Fatal("expected active service map, got nil")
	}
	if _, ok := active["5473"]; !ok {
		t.Fatalf("expected service 5473 to be active on 20261226")
	}
}

func TestBuildTrips(t *testing.T) {
	tripMap := trips.BuildTrips("4-VIC", "20261226")
	if len(tripMap) == 0 {
		t.Fatal("expected trips for 4-VIC on 20261226, got none")
	}

	trip, ok := tripMap["12337295:14530260:14827853"]
	if !ok {
		t.Fatal("expected known trip 12337295:14530260:14827853 in results")
	}
	if trip.RouteID != "4-VIC" {
		t.Fatalf("expected RouteID 4-VIC, got %q", trip.RouteID)
	}
	if trip.ServiceID != "5473" {
		t.Fatalf("expected ServiceID 5473, got %q", trip.ServiceID)
	}
	if trip.TripHeadsign == "" {
		t.Fatal("expected TripHeadsign to be set")
	}
	if len(trip.Stops) == 0 {
		t.Fatal("expected trip to have stops attached")
	}

	for i := 1; i < len(trip.Stops); i++ {
		if trip.Stops[i-1].StopSequence > trip.Stops[i].StopSequence {
			t.Fatalf("stops not sorted by sequence at index %d", i)
		}
	}

	first := trip.Stops[0]
	if first.StopID == "" || first.ArrivalTime == "" {
		t.Fatalf("expected first stop to have id and arrival time, got %+v", first)
	}
}
