# daylight [![Go Reference](https://pkg.go.dev/badge/github.com/lukasschwab/daylight.svg)](https://pkg.go.dev/github.com/lukasschwab/daylight)

`daylight` is a little macOS status bar app for telling you how much daylight remains.

+ Check a status bar countdown of how much daylight is left.
+ Create a calendar event for the last `n` minutes of daylight.

Due credit: `daylight` uses [progrium/macdriver](https://github.com/progrium/macdriver)'s Go bindings for Cocoa APIs.

![Screenshot of Daylight, showing the time remaining before sunset and a submenu for creating calendar invites.](./assets/screenshot.png)

## Usage

1. Download the latest version of Daylight.zip from [Releases](https://github.com/lukasschwab/daylight/releases).
2. Unzip Daylight.zip; this yields Daylight.app, the application bundle.
3. Move Daylight.app into your Applications folder.

### Changing the location

Daylight currently hardcodes coordinates for sunrise/sunset in San Francisco. If you'd like to use Daylight in some other city, change those hardcoded coordinates and rebuild the app:

1. Clone or download this repository.
2. Find the definitions of `cityLatitude` and `cityLongitude` in [data.go](./data.go). Update these to match the rough latitude and longitude of the city you want.<br>For example, if you're in Kansas City, you should have `cityLatitude = 39.0997` and `cityLongitude = -94.5786`.
3. Run `make install` to build Daylight and move it into your Applications folder.

### Development

The [Makefile](./Makefile) offers standard targets for building, running, and bundling Daylight. There are two make targets of particular interest for debugging:

+ `make dev` includes `cmd/daylight/dev.go` in the build. That causes logs to be double-written, to `stdout` and to the specified log file. See Go's [Build constraints](https://golang.org/cmd/go/#hdr-Build_constraints) docs for an explanation of how this works.
+ `make run` runs a dev binary.

`make app` composes a macOS application bundle, and `make install` copies that bundle into your `/Applications` directory (overwriting any existing installation).
