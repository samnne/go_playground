package gtfs

// CSV header schemas for GTFS files in data/.

// AgencyHeaders are the columns in data/agency.txt.
var AgencyHeaders = []string{
	"agency_id",
	"agency_name",
	"agency_url",
	"agency_timezone",
	"agency_lang",
	"agency_phone",
	"agency_fare_url",
	"agency_email",
}

// CalendarDatesHeaders are the columns in data/calendar_dates.txt.
var CalendarDatesHeaders = []string{
	"service_id",
	"date",
	"exception_type",
}

// FeedInfoHeaders are the columns in data/feed_info.txt.
var FeedInfoHeaders = []string{
	"feed_publisher_name",
	"feed_publisher_url",
	"feed_lang",
	"feed_version",
	"feed_start_date",
	"feed_end_date",
}

// RoutesHeaders are the columns in data/routes.txt.
var RoutesHeaders = []string{
	"route_id",
	"route_short_name",
	"route_long_name",
	"route_type",
	"route_color",
	"route_text_color",
}

// ShapesHeaders are the columns in data/shapes.txt.
var ShapesHeaders = []string{
	"shape_id",
	"shape_pt_lat",
	"shape_pt_lon",
	"shape_pt_sequence",
	"shape_dist_traveled",
}

// StopsHeaders are the columns in data/stops.txt.
var StopsHeaders = []string{
	"stop_id",
	"stop_name",
	"stop_lat",
	"stop_lon",
	"wheelchair_boarding",
	"stop_code",
}

// StopTimesHeaders are the columns in data/stop_times.txt.
var StopTimesHeaders = []string{
	"trip_id",
	"arrival_time",
	"departure_time",
	"stop_id",
	"stop_sequence",
	"shape_dist_traveled",
	"stop_headsign",
	"pickup_type",
	"drop_off_type",
	"timepoint",
}

// TripsHeaders are the columns in data/trips.txt.
var TripsHeaders = []string{
	"route_id",
	"service_id",
	"trip_id",
	"trip_headsign",
	"shape_id",
	"block_id",
	"direction_id",
}

// Agency represents a row in data/agency.txt.
type Agency struct {
	AgencyID       string `csv:"agency_id"`
	AgencyName     string `csv:"agency_name"`
	AgencyURL      string `csv:"agency_url"`
	AgencyTimezone string `csv:"agency_timezone"`
	AgencyLang     string `csv:"agency_lang"`
	AgencyPhone    string `csv:"agency_phone"`
	AgencyFareURL  string `csv:"agency_fare_url"`
	AgencyEmail    string `csv:"agency_email"`
}

// CalendarDate represents a row in data/calendar_dates.txt.
type CalendarDate struct {
	ServiceID     string `csv:"service_id"`
	Date          string `csv:"date"`
	ExceptionType string `csv:"exception_type"`
}

// FeedInfo represents a row in data/feed_info.txt.
type FeedInfo struct {
	FeedPublisherName string `csv:"feed_publisher_name"`
	FeedPublisherURL  string `csv:"feed_publisher_url"`
	FeedLang          string `csv:"feed_lang"`
	FeedVersion       string `csv:"feed_version"`
	FeedStartDate     string `csv:"feed_start_date"`
	FeedEndDate       string `csv:"feed_end_date"`
}

// Route represents a row in data/routes.txt.
type Route struct {
	RouteID        string `csv:"route_id"`
	RouteShortName string `csv:"route_short_name"`
	RouteLongName  string `csv:"route_long_name"`
	RouteType      string `csv:"route_type"`
	RouteColor     string `csv:"route_color"`
	RouteTextColor string `csv:"route_text_color"`
}

// Shape represents a row in data/shapes.txt.
type Shape struct {
	ShapeID           string `csv:"shape_id"`
	ShapePtLat        string `csv:"shape_pt_lat"`
	ShapePtLon        string `csv:"shape_pt_lon"`
	ShapePtSequence   string `csv:"shape_pt_sequence"`
	ShapeDistTraveled string `csv:"shape_dist_traveled"`
}

// Stop represents a row in data/stops.txt.
type Stop struct {
	StopID           string `csv:"stop_id"`
	StopName         string `csv:"stop_name"`
	StopLat          string `csv:"stop_lat"`
	StopLon          string `csv:"stop_lon"`
	WheelchairBoarding string `csv:"wheelchair_boarding"`
	StopCode         string `csv:"stop_code"`
}

// StopTime represents a row in data/stop_times.txt.
type StopTime struct {
	TripID            string `csv:"trip_id"`
	ArrivalTime       string `csv:"arrival_time"`
	DepartureTime     string `csv:"departure_time"`
	StopID            string `csv:"stop_id"`
	StopSequence      string `csv:"stop_sequence"`
	ShapeDistTraveled string `csv:"shape_dist_traveled"`
	StopHeadsign      string `csv:"stop_headsign"`
	PickupType        string `csv:"pickup_type"`
	DropOffType       string `csv:"drop_off_type"`
	Timepoint         string `csv:"timepoint"`
}

// Trip represents a row in data/trips.txt.
type Trip struct {
	RouteID      string `csv:"route_id"`
	ServiceID    string `csv:"service_id"`
	TripID       string `csv:"trip_id"`
	TripHeadsign string `csv:"trip_headsign"`
	ShapeID      string `csv:"shape_id"`
	BlockID      string `csv:"block_id"`
	DirectionID  string `csv:"direction_id"`
}

// FileHeaders maps GTFS data filenames to their CSV headers.
var FileHeaders = map[string][]string{
	"agency.txt":         AgencyHeaders,
	"calendar_dates.txt": CalendarDatesHeaders,
	"feed_info.txt":      FeedInfoHeaders,
	"routes.txt":         RoutesHeaders,
	"shapes.txt":         ShapesHeaders,
	"stops.txt":          StopsHeaders,
	"stop_times.txt":     StopTimesHeaders,
	"trips.txt":          TripsHeaders,
}
