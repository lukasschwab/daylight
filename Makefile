BUILDDIR = build

# Build a production binary.
.PHONY: build
build: $(BUILDDIR)/daylight

$(BUILDDIR)/daylight: *.go cmd/daylight/*.go
	go build -o $(BUILDDIR)/daylight ./cmd/daylight

# Build a dev binary.
.PHONY: dev
dev:
	go build -tags dev -o $(BUILDDIR)/daylight ./cmd/daylight

# Run a dev binary.
.PHONY: run
run: dev
	./$(BUILDDIR)/daylight

# Bundle a production binary into a MacOS app.
.PHONY: app
app: $(BUILDDIR)/Daylight.app

$(BUILDDIR)/Daylight.app: build assets/*
	# Copy assets and build binaries into app directory structure.
	mkdir -p $(BUILDDIR)/Daylight.app/Contents/MacOS
	cp ./assets/Info.plist $(BUILDDIR)/Daylight.app/Contents
	cp $(BUILDDIR)/daylight $(BUILDDIR)/Daylight.app/Contents/MacOS
	mkdir -p $(BUILDDIR)/Daylight.app/Contents/Resources
	cp ./assets/icon.icns $(BUILDDIR)/Daylight.app/Contents/Resources

# Throw the built app in your /Applications directory.
.PHONY: install
install: app
	cp -r $(BUILDDIR)/Daylight.app /Applications

# Zip a production MacOS app.
.PHONY: zip
zip: $(BUILDDIR)/Daylight.zip

$(BUILDDIR)/Daylight.zip: $(BUILDDIR)/Daylight.app
	cd $(BUILDDIR) && zip -r Daylight.zip Daylight.app

.PHONY: clean
clean:
	rm -rf $(BUILDDIR)