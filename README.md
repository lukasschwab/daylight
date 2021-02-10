# daylight

`daylight` is a little app for telling you how much daylight remains, using data from [wttr.in](https://github.com/chubin/wttr.in).

## Features

+ During the day, a little status bar indication of how much daylight is left.
+ Create a calendar event for the last `n` minutes of daylight.

## To do

- [ ] Reasonable error handling.
- [ ] Get a proper app icon.
- [ ] Don't hardcode San Francisco.
- [ ] Interactions.
    - [ ] Figure out a notification flow.
    - [ ] Find a nice default representation for the status bar.
- [ ] Build.
    - [ ] Move build artifacts into a `build` directory for cleanliness.
    - [ ] `make` targets shouldn't all be phony.