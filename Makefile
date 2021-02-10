build:
	go build ./cmd/daylight

run:
	go run ./cmd/daylight/main.go

app: clean build
	mkdir -p ./Daylight.app/Contents/MacOS
	cp ./assets/Info.plist ./Daylight.app/Contents
	cp ./daylight ./Daylight.app/Contents/MacOS
	mkdir -p ./Daylight.app/Contents/Resources
	cp ./assets/icon.icns ./Daylight.app/Contents/Resources

zip: app
	zip -r Daylight.zip ./Daylight.app

clean:
	rm -f daylight
	rm -rf Daylight.app
	rm -f Daylight.zip