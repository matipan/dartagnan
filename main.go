package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"image"

	"flag"

	"github.corp.globant.com/InternetOfThings/face-tracking-turret/motion"
	"gocv.io/x/gocv"
)

const (
	minArea = 7000
)

var (
	stream = flag.Bool("stream", true, "stream the images on a server")
	port   = flag.String("port", ":8080", "port for the stream server")
	area   = flag.Float64("area", minArea, "base area for motion detection")
	device = flag.Int("device", 0, "device ID for the camera")
)

func main() {
	flag.Parse()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	s := motion.NewServer()
	detector, err := motion.NewDetector(*device, *area, contourFunc, s)
	if err != nil {
		log.Fatalf("Could not initialize detector: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go detector.Run(ctx)

	quit := make(chan struct{})
	go func() {
		<-c
		cancel()
		close(quit)
	}()

	if !*stream {
		<-quit
		log.Println("Leaving")
		return
	}

	server := &http.Server{
		Handler: s,
		Addr:    *port,
	}
	go func() {
		<-quit
		server.Shutdown(ctx)
	}()
	log.Println(server.ListenAndServe())
}

func contourFunc(rect image.Rectangle, frame gocv.Mat) {
	log.Printf("Got motion")
}
