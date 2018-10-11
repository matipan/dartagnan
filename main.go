package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"math"

	"image"
	"image/color"

	"flag"

	"gobot.io/x/gobot/platforms/raspi"
	"gobot.io/x/gobot/sysfs"
	"github.corp.globant.com/InternetOfThings/face-tracking-turret/motion"
	"gocv.io/x/gocv"
)

const (
	minArea = 7000

	angleMax = 2350000
	angleMin = 450000

	grids = 7

	distance float64 = 1.35

	minX = 70
	minY = 13
)

var (
	servoX sysfs.PWMPinner
	servoY sysfs.PWMPinner

	colorBlue = color.RGBA{B: 255}

	lastRun int64
	lastX, lastY, midX, midY float64
)

func calcDutyCycle(base uint8) uint32 {
	if base > 180 {
		base = 180
	}
	angle := uint32((float32(base)/180)*(angleMax-angleMin))+angleMin
	log.Printf("Angle: %d", angle)
	return angle
}

func moveX(angle uint8) {
	servoX.SetDutyCycle(calcDutyCycle(angle))
}

func moveY(angle uint8) {
	servoY.SetDutyCycle(calcDutyCycle(angle))
}

var (
	stream = flag.Bool("stream", true, "stream the images on a server")
	port   = flag.String("port", ":8080", "port for the stream server")
	area   = flag.Float64("area", minArea, "base area for motion detection")
	device = flag.Int("device", 0, "device ID for the camera")
)

type window struct {
	frame *gocv.Window
	delta *gocv.Window
	thresh *gocv.Window
}

func (w *window) StreamFrame(img gocv.Mat) {
	gocv.Resize(img, &img, image.Point{X: 800, Y: 600}, 0, 0, gocv.InterpolationLinear)
	w.frame.IMShow(img)
	w.frame.WaitKey(1)
}

func (w *window) StreamThresh(img gocv.Mat) {
	w.thresh.IMShow(img)
	w.thresh.WaitKey(1)
}

func (w *window) StreamDelta(img gocv.Mat) {
	w.delta.IMShow(img)
	w.delta.WaitKey(1)
}

func main() {
	flag.Parse()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// s := motion.NewServer()
	frames := gocv.NewWindow("Frame")
	deltas := gocv.NewWindow("Delta")
	threshs := gocv.NewWindow("Threshs")
	frames.ResizeWindow(800, 600)
	deltas.ResizeWindow(400,400)
	threshs.ResizeWindow(400,400)
	frames.MoveWindow(0, 0)
	deltas.MoveWindow(800, 0)
	threshs.MoveWindow(800, 600)
	w := &window{frame: frames, delta: deltas, thresh: threshs}
	detector, err := motion.NewDetector(*device, *area, contourFunc, w)
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

	r := raspi.NewAdaptor()
	r.PiBlasterPeriod = 20000000
	r.Connect()
	servoX, err = r.PWMPin("33")
	if err != nil {
		log.Fatal(err)
	}
	servoY, err = r.PWMPin("35")
	if err != nil {
		log.Fatal(err)
	}
	moveX(0)
	moveY(0)

	<-quit
	w.frame.Close()
	w.delta.Close()
	w.thresh.Close()
}

func contourFunc(rect image.Rectangle, img gocv.Mat) {
	now := time.Now().Unix()*1000
	if (now - lastRun) < 10 {
		return
	}
	midX = (float64((rect.Max.X-rect.Min.X))/2.0+float64(rect.Min.X))
	midY = (float64((rect.Max.Y-rect.Min.Y))/2.0+float64(rect.Min.Y))
	if lastX == midX && lastY == midY {
		return
	}
	lastX = midX
	lastY = midY
	radsX := math.Atan((midX*distance)/500.0)
	radsY := math.Atan(((500.0-midY)*distance)/500.0)
	x := uint8(radsX*180/math.Pi)+minX
	y := uint8(radsY*180/math.Pi)-minY

	log.Printf("pixels(x,y)=(%v,%v) -- angles(x,y)=(%v,%v) -- rads(x,y)=(%.4f,%.4f)", int(midX), int(midY), x, y, radsX, radsY)
	moveY(y)
	moveX(x)
	lastRun = now
}
