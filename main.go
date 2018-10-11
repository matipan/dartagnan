package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"flag"

	"github.corp.globant.com/InternetOfThings/face-tracking-turret/detector"
	"github.corp.globant.com/InternetOfThings/face-tracking-turret/turret"
	"github.corp.globant.com/InternetOfThings/face-tracking-turret/window"
)

const (
	defaultArea   = 7000
	defaultDevice = 0
)

var (
	area   = flag.Float64("area", defaultArea, "base area for motion detection")
	device = flag.Int("device", defaultDevice, "device ID for the camera")
)

func main() {
	flag.Parse()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	t, err := turret.New("35", "33", 1.35, 500, 50)
	if err != nil {
		log.Printf("Could not create turret: %s", err)
		os.Exit(1)
	}

	wm := window.New(800, 600)
	detector, err := detector.New(*device, *area, t.HandleMotion, wm)
	if err != nil {
		log.Printf("Could not initialize detector: %s", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go detector.Run(ctx)
	<-c
	cancel()
	wm.Close()
	close(c)
}
