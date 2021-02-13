package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/laurent22/ical-go/ical"
	"github.com/lukasschwab/daylight"

	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/objc"
)

func main() {
	runtime.LockOSThread()
	app := cocoa.NSApp_WithDidLaunch(func(n objc.Object) {
		// Track temporary ICS files and clean them up on close.
		eventTempFiles := &daylight.TempFiles{FileNameFormat: "daylight.*.ics"}
		defer eventTempFiles.CleanUp()

		// Make channels for handling user clicks.
		refreshClicked := make(chan bool)   // channels manual refresh triggers.
		newEventClicked := make(chan int64) // channels requests for calendar events.
		// Initialize UI.
		ui := daylight.InitUI(refreshClicked, newEventClicked)

		go func() {
			// fetchedData is cached sunrise/sunset data; it's expected to last
			// for a day before it's automatically refetched.
			var fetchedData *daylight.SunData
			fetchAndRender := func(forceRefetch bool) {
				if forceRefetch {
					ui.SetStatusItemTitle(daylight.TitleLoading)
				}
				var err error
				if fetchedData, err = fetchedData.Update(forceRefetch); err != nil {
					log.Printf("Error updating data: %v", err)
				}
				ui.Render(fetchedData)
			}
			// Initialize state.
			fetchAndRender(true)

			// Event loop.
			for {
				select {
				case <-time.After(1 * time.Minute):
					fetchAndRender(false)
				case <-time.After(15 * time.Minute):
					// Periodically lean up created event files.
					eventTempFiles.CleanUp()
				case <-refreshClicked:
					fetchAndRender(true)
				case minutes := <-newEventClicked:
					openICSEvent(eventTempFiles, fetchedData.Sunset, minutes)
				}
			}
		}()
	})
	app.Run()
}

// openICSEvent writes an ICS calendar event of the specified duration (ending
// at sunset) to a temporary file, then opens that file with the default app.
func openICSEvent(tmpfiles *daylight.TempFiles, sunset time.Time, minutes int64) {
	log.Printf("Creating a %d-minute calendar event\n", minutes)
	// Fill out an ICS event.
	startAt := sunset.Add(time.Duration(minutes) * -time.Minute)
	calendar := ical.Calendar{Items: []ical.CalendarEvent{{
		Id:      uuid.New().String(),
		Summary: fmt.Sprintf("☀️ %d minutes to sunset", minutes),
		StartAt: &startAt,
		EndAt:   &sunset,
	}}}

	// Write temporary file.
	icsEventFile, err := tmpfiles.New()
	if err != nil {
		log.Printf("Encountered error creating temporary ICS file: %v\n", err)
		return
	}
	icsEventFile.Write([]byte(calendar.ToICS()))
	icsEventFile.Close()

	// Open temporary file.
	cmd := exec.Command("open", icsEventFile.Name())
	if err := cmd.Run(); err != nil {
		log.Printf("Encountered error opening ICS file %v: %v\n", icsEventFile.Name(), err)
	}
}
