package daylight

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	// TODO: Parameterize location or use default location-finding.
	// wttrURL yields a response of the form {"sunrise":"07:04:45","sunset":"17:43:31"}.
	wttrURL = `https://wttr.in/San+Francisco?format={"sunrise":"%S","sunset":"%s"}`

	// rawTimeLayout corresponds to the time format returned by wttr.in.
	rawTimeLayout = "15:04:05"
)

// SunData represents sunlight data for a given day at a given location.
type SunData struct {
	// fetchTime is the rough time when this SunData was fetched from wttr.in.
	fetchTime time.Time
	Sunrise   time.Time
	Sunset    time.Time
}

// GetCurrentData fetches sunrise and sunset data from wttr.in.
func GetCurrentData() (*SunData, error) {
	log.Printf("Fetching %v", wttrURL)
	resp, err := http.Get(wttrURL)
	if err != nil {
		return nil, fmt.Errorf("Error fetching wttr.in data: %w", err)
	}
	rawData := &rawSunData{}
	if err = json.NewDecoder(resp.Body).Decode(rawData); err != nil {
		return nil, fmt.Errorf("Error decoding JSON response: %w", err)
	}
	return rawData.convert()
}

// NeedsRefresh returns true if d was fetched on a day before today.
func (d *SunData) NeedsRefresh() bool {
	return time.Now().Day() != d.fetchTime.Day()
}

// rawSunData corresponds to the raw data returned by wttr.in. See wttrURL.
type rawSunData struct {
	Sunrise string `json:"sunrise"`
	Sunset  string `json:"sunset"`
}

// convert parses rawSunData times into SunData times.
func (rd *rawSunData) convert() (*SunData, error) {
	// FIXME: don't hardcode tz. Also, handle error.
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return nil, err
	}
	now := time.Now()
	year, month, day := now.Date()

	parse := func(rawTime string) (time.Time, error) {
		parsed, err := time.Parse(rawTimeLayout, rawTime)
		if err != nil {
			return parsed, fmt.Errorf("Error parsing raw time: %w", err)
		}
		return time.Date(year, month, day, parsed.Hour(), parsed.Minute(), parsed.Second(), 0, location), nil
	}

	sunrise, err := parse(rd.Sunrise)
	if err != nil {
		return nil, err
	}
	sunset, err := parse(rd.Sunset)
	if err != nil {
		return nil, err
	}
	return &SunData{now, sunrise, sunset}, nil
}
