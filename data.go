package daylight

import (
	"time"

	"github.com/lukasschwab/sunrisesunset"
)

const (
	// Coordinates of San Francisco, CA. TODO: parameterize.
	cityLatitude  = 37.7749
	cityLongitude = -122.4194
)

// SunData represents sunlight data for a given day at a given location.
type SunData struct {
	// calculatedAt is the rough time when this SunData was calculated.
	calculatedAt    time.Time
	Sunrise         time.Time
	Sunset          time.Time
	SunriseTomorrow time.Time
}

// Update recalculates the current sunrise/sunset if it was last calculated
// before today.
func (d *SunData) Update() (data *SunData, err error) {
	now := time.Now()
	// If the existing data is up-to-date, no need to recalculate.
	if d != nil && d.calculatedAt.Day() == now.Day() {
		return d, nil
	}

	// Calculate sunrise/sunset times.
	sunrise, sunset, err := getSunriseSunset(now)
	if err != nil {
		return nil, err
	}
	tomorrowSunrise, _, err := getSunriseSunset(now.Add(24 * time.Hour))
	if err != nil {
		return nil, err
	}
	return &SunData{now, sunrise, sunset, tomorrowSunrise}, nil
}

// getSunriseSunset calculates and normalizes the sunrise and sunset times on
// the day, and in the time zone, specified by t.
func getSunriseSunset(t time.Time) (sunrise, sunset time.Time, err error) {
	params := sunrisesunset.Parameters{
		Latitude:  cityLatitude,
		Longitude: cityLongitude,
		UtcOffset: getOffset(t),
		Date:      t,
	}
	return params.GetSunriseSunset()
}

// getOffset converts a time's UTC offset (provided in seconds) into the format
// accepted by sunrisesunset (a float64 number of hours).
func getOffset(t time.Time) float64 {
	_, offsetSeconds := t.Zone()
	return float64(offsetSeconds) / time.Hour.Seconds()
}
