package main

import (
	"log"
	"strconv"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/luismesas/goPi/piface"
	"github.com/luismesas/goPi/spi"
)

var pfd *piface.PiFaceDigital

func init() {
	pfd = piface.NewPiFaceDigital(spi.DEFAULT_HARDWARE_ADDR, spi.DEFAULT_BUS, spi.DEFAULT_CHIP)
	err := pfd.InitBoard()
	if err != nil {
		log.Fatal(err)
	}
}

func unlockDoorForSeconds(secs int) {
	sString := strconv.Itoa(secs)
	log15.Info("Opening door for " + sString + " seconds.")
	pfd.OutputPins[1].AllOn()
	afterChan := time.After(time.Duration(secs) * time.Second)
	go func(when <-chan time.Time) {
		<-when
		log15.Info("Locking door.")
		// TODO: Verify this is indempotent and won't crash if
		// two clients' calls overlap.
		pfd.OutputPins[1].AllOff()
	}(afterChan)
}

func testUnlockingFunction(secs int) {
	sString := strconv.Itoa(secs)
	log15.Info("Opening door for " + sString + " seconds.")
	afterChan := time.After(time.Duration(secs) * time.Second)
	go func(when <-chan time.Time) {
		<-when
		log15.Info("Locking door.")
	}(afterChan)
}
