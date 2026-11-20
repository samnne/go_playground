package trips

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// LoadGTFS loads GTFS data into memory once at startup.
// It accepts a path to a data directory or a .zip file. If the path is a
// directory and not usable, it will try data.zip as a fallback.
func LoadGTFS(path string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.activeServices = make(map[string]map[string]struct{})
	store.tripsByRoute = make(map[string][]*Trip)
	store.tripsByID = make(map[string]*Trip)
	store.routes = make(map[string]*Route)
	store.loaded = false

	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		if err := loadFromDir(path); err != nil {
			return err
		}
		store.loaded = true
		return nil
	}
	if err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".zip") {
		if err := loadFromZip(path); err != nil {
			return err
		}
		store.loaded = true
		return nil
	}

	// Fallbacks: try data.zip, then data directory.
	if _, zerr := os.Stat("data.zip"); zerr == nil {
		if err := loadFromZip("data.zip"); err != nil {
			return err
		}
		store.loaded = true
		return nil
	}
	if derr := loadFromDir("data"); derr == nil {
		store.loaded = true
		return nil
	} else if err != nil {
		return err
	} else {
		return derr
	}
}

func loadFromDir(dataDir string) error {
	// 1. Calendar dates
	if f, err := os.Open(filepath.Join(dataDir, "calendar_dates.txt")); err == nil {
		if err := readCalendarDates(f); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	// 2. Stops
	stopDetailsMap := make(map[string]stopDetails)
	if f, err := os.Open(filepath.Join(dataDir, "stops.txt")); err == nil {
		details, err := readStopDetails(f)
		f.Close()
		if err != nil {
			return err
		}
		stopDetailsMap = details
	}

	// 2b. Routes
	if f, err := os.Open(filepath.Join(dataDir, "routes.txt")); err == nil {
		if err := readRoutes(f); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	// 3. Trips
	tf, err := os.Open(filepath.Join(dataDir, "trips.txt"))
	if err != nil {
		return err
	}
	if err := readTrips(tf); err != nil {
		tf.Close()
		return err
	}
	tf.Close()

	// 4. Stop times
	if f, err := os.Open(filepath.Join(dataDir, "stop_times.txt")); err == nil {
		if err := readStopTimes(f, stopDetailsMap); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	sortAllTripStops()
	return nil
}

func loadFromZip(zipPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	files := make(map[string]*zip.File, len(r.File))
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := filepath.Base(f.Name)
		files[name] = f
	}

	// 1. Calendar dates
	if zf, ok := files["calendar_dates.txt"]; ok {
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		if err := readCalendarDates(rc); err != nil {
			rc.Close()
			return err
		}
		rc.Close()
	}

	// 2. Stops
	stopDetailsMap := make(map[string]stopDetails)
	if zf, ok := files["stops.txt"]; ok {
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		details, err := readStopDetails(rc)
		rc.Close()
		if err != nil {
			return err
		}
		stopDetailsMap = details
	}

	// 2b. Routes
	if zf, ok := files["routes.txt"]; ok {
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		if err := readRoutes(rc); err != nil {
			rc.Close()
			return err
		}
		rc.Close()
	}

	// 3. Trips
	zf, ok := files["trips.txt"]
	if !ok {
		return fmt.Errorf("trips.txt not found in zip: %s", zipPath)
	}
	rc, err := zf.Open()
	if err != nil {
		return err
	}
	if err := readTrips(rc); err != nil {
		rc.Close()
		return err
	}
	rc.Close()

	// 4. Stop times
	if zf, ok := files["stop_times.txt"]; ok {
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		if err := readStopTimes(rc, stopDetailsMap); err != nil {
			rc.Close()
			return err
		}
		rc.Close()
	}

	sortAllTripStops()
	return nil
}

func sortAllTripStops() {
	for _, t := range store.tripsByID {
		sort.Slice(t.Stops, func(i, j int) bool {
			return t.Stops[i].StopSequence < t.Stops[j].StopSequence
		})
	}
}

func attachStopsLegacy(trips map[string]*Trip) {
	if len(trips) == 0 {
		return
	}
	stops := loadStopDetailsLegacy()
	file, err := os.Open("data/stop_times.txt")
	if err != nil {
		return
	}
	idx, reader := openCSVReaderLegacy(file, err)
	defer file.Close()
	if reader == nil {
		return
	}

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

		trip, ok := trips[record[tripCol]]
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
		sd := stops[stopID]
		trip.Stops = append(trip.Stops, Stop{
			StopID:       stopID,
			StopName:     sd.name,
			ArrivalTime:  record[arrCol],
			StopSequence: seq,
			StopLat:      sd.lat,
			StopLong:     sd.long,
		})
	}

	for _, trip := range trips {
		s := trip.Stops
		sort.Slice(s, func(a, b int) bool { return s[a].StopSequence < s[b].StopSequence })
	}
}

func loadStopDetailsLegacy() map[string]stopDetails {
	file, err := os.Open("data/stops.txt")
	if err != nil {
		return map[string]stopDetails{}
	}
	defer file.Close()
	details, err := readStopDetails(file)
	if err != nil {
		return map[string]stopDetails{}
	}
	return details
}
