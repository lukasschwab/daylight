BUILDDIR = build

.PHONY: build
build: $(BUILDDIR)/daylight

$(BUILDDIR)/daylight: *.go cmd/daylight/*.go
	go build -o $(BUILDDIR)/daylight ./cmd/daylight

.PHONY: run
run:
	go run ./cmd/daylight/main.go

.PHONY: app
app: $(BUILDDIR)/Daylight.app

$(BUILDDIR)/Daylight.app: $(BUILDDIR)/daylight assets/*
	# Copy assets and build binaries into app directory structure.
	mkdir -p $(BUILDDIR)/Daylight.app/Contents/MacOS
	cp ./assets/Info.plist $(BUILDDIR)/Daylight.app/Contents
	cp $(BUILDDIR)/daylight $(BUILDDIR)/Daylight.app/Contents/MacOS
	mkdir -p $(BUILDDIR)/Daylight.app/Contents/Resources
	cp ./assets/icon.icns $(BUILDDIR)/Daylight.app/Contents/Resources

.PHONY: zip
zip: $(BUILDDIR)/Daylight.zip

$(BUILDDIR)/Daylight.zip: $(BUILDDIR)/Daylight.app
	zip -r $(BUILDDIR)/Daylight.zip $(BUILDDIR)/Daylight.app

.PHONY: clean
clean:
	rm -rf $(BUILDDIR)