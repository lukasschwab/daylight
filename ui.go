package daylight

import (
	"fmt"
	"log"
	"time"

	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/objc"
)

const (
	// TitleLoading is used between launch and rendering of the first fetch.
	// It's the one title that may be set outside of Render().
	TitleLoading = "◌"
	// titleDark indicates the present time is before sunrise or after sunset.
	titleDark = "◻"
	// titleDaylightFormat indicates the present time is between sunrise and
	// sunset; it formats the duration until sunset. See: toString.
	titleDaylightFormat = "◼ %v"
)

// ui controls the representation of sunset data. This struct only includes
// references to those components which are functions of the sun data, but
// InitUI initializes the full UI.
type ui struct {
	// statusItem is the top level status bar icon.
	statusItem cocoa.NSStatusItem
	// verboseItem is the top menu item, which shows time to sunset.
	verboseItem cocoa.NSMenuItem
}

// InitUI initialized the full UI, and returns a ui struct which can be used
// to render updates to the data.
func InitUI(refreshClicked chan bool, newEventClicked chan int64) (components *ui) {
	// Initialize the stored components, which are functions of the sun data.

	// + root status bar icon
	statusItem := cocoa.NSStatusBar_System().StatusItemWithLength(cocoa.NSVariableStatusItemLength)
	statusItem.Retain()
	statusItem.Button().SetTitle(TitleLoading)
	// + unclickable item expressing the duration
	verboseItem := cocoa.NSMenuItem_New()
	components = &ui{statusItem, verboseItem}

	// Initialize static components: not modified in Render().

	// + menu item to manually refresh data
	itemRefresh := cocoa.NSMenuItem_New()
	itemRefresh.SetTitle("Refresh data")
	itemRefresh.SetAction(objc.Sel("refresh:"))
	cocoa.DefaultDelegateClass.AddMethod("refresh:", func(_ objc.Object) {
		refreshClicked <- true
	})

	// + calendar events submenu
	calendarEventsItem := cocoa.NSMenuItem_New()
	calendarEventsItem.SetTitle("New calendar event...")
	calendarEventsMenu := cocoa.NSMenu_New()
	for _, mins := range []int64{30, 60, 90} {
		calendarEventsMenu.AddItem(makeCalendarEventItem(mins, newEventClicked))
	}
	calendarEventsItem.SetSubmenu(calendarEventsMenu)

	// + menu item to quit
	itemQuit := cocoa.NSMenuItem_New()
	itemQuit.SetTitle("Quit Daylight")
	itemQuit.SetAction(objc.Sel("terminate:"))

	// Assemble menu items into menu, then attach menu to status bar obj.
	menu := cocoa.NSMenu_New()
	menu.AddItem(verboseItem)
	menu.AddItem(itemRefresh)
	menu.AddItem(cocoa.NSMenuItem_Separator())
	menu.AddItem(calendarEventsItem)
	menu.AddItem(cocoa.NSMenuItem_Separator())
	menu.AddItem(itemQuit)
	statusItem.SetMenu(menu)

	return components
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

// Render updates components to reflect data.
func (components *ui) Render(data *SunData) {
	now := time.Now()
	if data == nil {
		// Unrenderable state; don't change current
		log.Println("Tried rendering nil data")
	} else if now.Before(data.Sunrise) {
		// Indicate waiting for sunrise.
		components.SetStatusItemTitle(titleDark)
		toSunrise := data.Sunrise.Sub(now).Round(time.Minute)
		components.verboseItem.SetTitle(fmt.Sprintf("%v until sunrise", toSunrise.String()))
	} else if now.After(data.Sunset) {
		// Indicate no data for tomorrow.
		components.SetStatusItemTitle(titleDark)
		components.verboseItem.SetTitle("You snooze, you lose.")
	} else {
		// Indicate time to sunset.
		toSunset := data.Sunset.Sub(now).Round(time.Minute)
		toSunsetString := toString(toSunset)
		components.SetStatusItemTitle(fmt.Sprintf(titleDaylightFormat, toSunsetString))
		components.verboseItem.SetTitle(fmt.Sprintf("%v until sunset", toSunsetString))
	}
}

// SetStatusItemTitle sets the title for the status bar item. This setter is
// exported as an exception:
func (components *ui) SetStatusItemTitle(title string) {
	components.statusItem.Button().SetTitle(title)
}

// toString formats a duration until sunset for display in the status bar and
// in the verbose menu item.
func toString(d time.Duration) string {
	hours := d / time.Hour
	minutes := (d - (hours * time.Hour)) / time.Minute
	return fmt.Sprintf("%dh%dm", hours, minutes)
}
