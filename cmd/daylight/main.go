package main

// TODO: refactor logic into a daylight package in the root, and refactor the application into ./cmd

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
		obj := cocoa.NSStatusBar_System().StatusItemWithLength(cocoa.NSVariableStatusItemLength)
		obj.Retain()
		obj.Button().SetTitle("🌓")

		itemVerbose := cocoa.NSMenuItem_New()
		eventTempFiles := &daylight.TempFiles{FileNameFormat: "daylight.*.ics"}
		defer eventTempFiles.CleanUp()

		refreshClicked := make(chan bool)
		go func() {
			// TODO: error handling.
			var fetchedData *daylight.SunData

			render := func() {
				now := time.Now()
				if now.Before(fetchedData.Sunrise) {
					obj.Button().SetTitle("◻")
					toSunrise := fetchedData.Sunrise.Sub(now).Round(time.Minute)
					itemVerbose.SetTitle(fmt.Sprintf("%v until sunrise", toSunrise.String()))
				} else if now.After(fetchedData.Sunset) {
					obj.Button().SetTitle("◻")
					itemVerbose.SetTitle("You snooze, you lose.")
				} else {
					toSunset := fetchedData.Sunset.Sub(now).Round(time.Minute)
					toSunsetString := toString(toSunset)
					obj.Button().SetTitle(fmt.Sprintf("◼ %v", toSunsetString))
					itemVerbose.SetTitle(fmt.Sprintf("%v until sunset", toSunsetString))
				}
			}

			// Initial state.
			fetchedData, _ = daylight.GetCurrentData()
			render()

			// Event loop.
			for {
				select {
				case <-time.After(1 * time.Minute):
					// Refetch data if necessary. TODO: error handling.
					if fetchedData.NeedsRefresh() {
						fetchedData, _ = daylight.GetCurrentData()
					}
					render()
				case <-time.After(15 * time.Minute):
					// Clean up created event files.
					eventTempFiles.CleanUp()
				case <-refreshClicked:
					fetchedData, _ = daylight.GetCurrentData()
					render()
				}
			}
		}()

		itemQuit := cocoa.NSMenuItem_New()
		itemQuit.SetTitle("Quit Daylight")
		itemQuit.SetAction(objc.Sel("terminate:"))

		itemRefresh := cocoa.NSMenuItem_New()
		itemRefresh.SetTitle("Refresh data")
		itemRefresh.SetAction(objc.Sel("refresh:"))
		cocoa.DefaultDelegateClass.AddMethod("refresh:", func(_ objc.Object) {
			refreshClicked <- true
		})

		calendarEventsItem := cocoa.NSMenuItem_New()
		calendarEventsItem.SetTitle("New calendar event...")
		calendarEventsMenu := cocoa.NSMenu_New()
		for _, mins := range []int64{30, 60, 90} {
			calendarEventsMenu.AddItem(makeCalendarEventItem(mins, eventTempFiles))
		}
		calendarEventsItem.SetSubmenu(calendarEventsMenu)

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

func makeCalendarEventItem(minutes int64, tmpfiles *daylight.TempFiles) cocoa.NSMenuItem {
	selector := fmt.Sprintf("calendar%d:", minutes)
	item := cocoa.NSMenuItem_New()
	item.SetTitle(fmt.Sprintf("Last %d minutes", minutes))
	item.SetAction(objc.Sel(selector))
	cocoa.DefaultDelegateClass.AddMethod(selector, func(_ objc.Object) {
		log.Printf("Creating a %d-minute calendar event\n", minutes)
		// Fill out an ICS event.
		// TODO: we can send a message via a channel to the main loop to do this from there and use their state...
		data, _ := daylight.GetCurrentData()
		startAt := data.Sunset.Add(time.Duration(minutes) * -time.Minute)
		calendar := ical.Calendar{Items: []ical.CalendarEvent{{
			Id:       uuid.New().String(),
			Summary:  fmt.Sprintf("☀️ %d minutes to sunset", minutes),
			Location: "San Francisco", // FIXME
			StartAt:  &startAt,
			EndAt:    &data.Sunset,
		}}}

		// Write temporary file.
		icsEventFile, err := tmpfiles.New()
		if err != nil {
			log.Printf("Encountered error creating temporary ICS file: %v", err)
			return
		}
		icsEventFile.Write([]byte(calendar.ToICS()))
		icsEventFile.Close()

		// Open temporary file.
		cmd := exec.Command("open", icsEventFile.Name())
		if err := cmd.Run(); err != nil {
			log.Printf("Encountered error opening ICS file %v: %v", icsEventFile.Name(), err)
		}
	})
	return item
}

func toString(d time.Duration) string {
	hours := d / time.Hour
	minutes := (d - (hours * time.Hour)) / time.Minute
	return fmt.Sprintf("%dh%dm", hours, minutes)
}
