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

const (
	// titleLoading is used between launch and rendering of the first fetch.
	titleLoading = "◌"
	// titleDark indicates the present time is before sunrise or after sunset.
	titleDark = "◻"
	// titleDaylightFormat indicates the present time is between sunrise and
	// sunset; it formats the duration until sunset. See: toString.
	titleDaylightFormat = "◼ %v"
)

func main() {
	log.Println("Starting Daytime")
	runtime.LockOSThread()

	app := cocoa.NSApp_WithDidLaunch(func(n objc.Object) {
		// Init status bar icon.
		obj := cocoa.NSStatusBar_System().StatusItemWithLength(cocoa.NSVariableStatusItemLength)
		obj.Retain()
		obj.Button().SetTitle(titleLoading)
		itemVerbose := cocoa.NSMenuItem_New()

		// Track temporary ICS files and clean them up on close.
		eventTempFiles := &daylight.TempFiles{FileNameFormat: "daylight.*.ics"}
		defer eventTempFiles.CleanUp()

		// Make channels for handling user clicks.
		refreshClicked := make(chan bool)   // channels manual refresh triggers.
		newEventClicked := make(chan int64) // channels requests for calendar events.

		render := func(data *daylight.SunData) {
			now := time.Now()
			if data == nil {
				// Unrenderable state; don't change current
				log.Println("Tried rendering nil data")
			} else if now.Before(data.Sunrise) {
				// Indicate waiting for sunrise.
				obj.Button().SetTitle(titleDark)
				toSunrise := data.Sunrise.Sub(now).Round(time.Minute)
				itemVerbose.SetTitle(fmt.Sprintf("%v until sunrise", toSunrise.String()))
			} else if now.After(data.Sunset) {
				// Indicate no data for tomorrow.
				obj.Button().SetTitle(titleDark)
				itemVerbose.SetTitle("You snooze, you lose.")
			} else {
				// Indicate time to sunset.
				toSunset := data.Sunset.Sub(now).Round(time.Minute)
				toSunsetString := toString(toSunset)
				obj.Button().SetTitle(fmt.Sprintf(titleDaylightFormat, toSunsetString))
				itemVerbose.SetTitle(fmt.Sprintf("%v until sunset", toSunsetString))
			}
		}

		go func() {
			// fetchedData is cached sunrise/sunset data; it's expected to last
			// for a day before it's automatically refetched.
			var fetchedData *daylight.SunData
			fetchAndRender := func(forceRefetch bool) {
				if forceRefetch {
					obj.Button().SetTitle(titleLoading)
				}
				var err error
				if fetchedData, err = fetchedData.Update(forceRefetch); err != nil {
					log.Printf("Error updating data: %v", err)
				}
				render(fetchedData)
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

		// Construct menu item to quit app.
		itemQuit := cocoa.NSMenuItem_New()
		itemQuit.SetTitle("Quit Daylight")
		itemQuit.SetAction(objc.Sel("terminate:"))

		// Construct menu item to manually refresh data.
		itemRefresh := cocoa.NSMenuItem_New()
		itemRefresh.SetTitle("Refresh data")
		itemRefresh.SetAction(objc.Sel("refresh:"))
		cocoa.DefaultDelegateClass.AddMethod("refresh:", func(_ objc.Object) {
			refreshClicked <- true
		})

		// Construct menu items to create calendar events.
		calendarEventsItem := cocoa.NSMenuItem_New()
		calendarEventsItem.SetTitle("New calendar event...")
		calendarEventsMenu := cocoa.NSMenu_New()
		for _, mins := range []int64{30, 60, 90} {
			calendarEventsMenu.AddItem(makeCalendarEventItem(mins, newEventClicked))
		}
		calendarEventsItem.SetSubmenu(calendarEventsMenu)

		// Assemble menu items into menu, then attach menu to status bar obj.
		menu := cocoa.NSMenu_New()
		menu.AddItem(itemVerbose)
		menu.AddItem(itemRefresh)
		menu.AddItem(cocoa.NSMenuItem_Separator())
		menu.AddItem(calendarEventsItem)
		menu.AddItem(cocoa.NSMenuItem_Separator())
		menu.AddItem(itemQuit)
		obj.SetMenu(menu)
	})
	app.Run()
}

// makeCalendarEventItem constructs an NSMenuItem which, when clicked, signals
// the event loop via ch to open an ICS event of the specified duration.
func makeCalendarEventItem(minutes int64, ch chan int64) cocoa.NSMenuItem {
	selector := fmt.Sprintf("calendar%d:", minutes)
	item := cocoa.NSMenuItem_New()
	item.SetTitle(fmt.Sprintf("Last %d minutes", minutes))
	item.SetAction(objc.Sel(selector))
	cocoa.DefaultDelegateClass.AddMethod(selector, func(_ objc.Object) {
		ch <- minutes
	})
	return item
}

// toString formats a duration until sunset for display in the status bar and
// in the verbose menu item.
func toString(d time.Duration) string {
	hours := d / time.Hour
	minutes := (d - (hours * time.Hour)) / time.Minute
	return fmt.Sprintf("%dh%dm", hours, minutes)
}

// openICSEvent writes an ICS calendar event of the specified duration (ending
// at sunset) to a temporary file, then opens that file with the default app.
func openICSEvent(tmpfiles *daylight.TempFiles, sunset time.Time, minutes int64) {
	log.Printf("Creating a %d-minute calendar event\n", minutes)
	// Fill out an ICS event.
	startAt := sunset.Add(time.Duration(minutes) * -time.Minute)
	calendar := ical.Calendar{Items: []ical.CalendarEvent{{
		Id:       uuid.New().String(),
		Summary:  fmt.Sprintf("☀️ %d minutes to sunset", minutes),
		Location: "San Francisco", // FIXME
		StartAt:  &startAt,
		EndAt:    &sunset,
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
