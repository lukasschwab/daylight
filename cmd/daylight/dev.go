// +build dev

package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const logFileName = "build/dev.log"

func init() {
	logFile, err := os.OpenFile(logFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Error opening or creating log file: %v", err)
	}
	// Write dev logs to file AND stdout.
	devWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(devWriter)

	log.Println("[ STARTING DAYTIME ]")

	// Close the file when the program is terminated. This is kind of a hack;
	// it's probably safe-ish to remove it if it's annoying.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		// Note: this code does *not* run if the user quits daylight via the
		// "Quit Daylight" menu item.
		log.Printf("[ CLOSING DAYTIME ]")
		logFile.Close()
		os.Exit(0)
	}()
}
