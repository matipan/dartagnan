package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.corp.globant.com/InternetOfThings/face-tracking-turret/server"
)

func main() {
	// TODO: flags with specific platform to run

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	s := &server.Server{}
	go func() {
		<-c
		s.Close()
	}()

	log.Fatal(s.Serve(":8080", "/", 0, "res10_300x300_ssd_iter_140000.caffemodel", "deploy.prototxt"))
}
