# daylight

`daylight` is a little app for telling you how much daylight remains, using data from [wttr.in](https://github.com/chubin/wttr.in).

## Features

+ During the day, a little status bar indication of how much daylight is left.
+ Create a calendar event for the last `n` minutes of daylight.

## To do

- [x] Reasonable error handling.
- [x] Encapsulate state management.
- [ ] Get a proper app icon.
- [ ] ~Don't hardcode San Francisco.~ I've done this as much as I can; the last challenge is wttr.in. Unfortunately, their automatic location detection seems to be broken, so there remains one place––in the consts in `data.go`––where San Francisco is hardcoded.
- [x] Interactions.
    - [x] Find a nice default representation for the status bar.
    - [ ] ~Figure out a notification flow.~ `beeep` notifications don't look great because they're with appscript. Notifications should be configurable, but I'm not inclined to manage a plist.
    - [ ] ~Surfacing errors to the user––not just in the logs.~ Moot: the app is more stable now, and I have a `dev` build flag for debugging.
- [x] Build.
    - [x] Move build artifacts into a `build` directory for cleanliness.
    - [x] `make` targets shouldn't all be phony.
- [ ] Usage instructions in README/a brief overview of the tech.