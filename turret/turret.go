package turret

import (
	"image"
	"log"
	"math"
	"time"

	"github.com/pkg/errors"
	"gobot.io/x/gobot/platforms/raspi"
	"gobot.io/x/gobot/sysfs"
)

const (
	dcMax = 2350000
	dcMin = 450000

	grids = 7

	distance float64 = 1.35

	minX = 70
	minY = 13
)

var (
	lastRun                  uint64
	lastX, lastY, midX, midY int
	x, y                     uint8
)

// Turret is the aiming turret that handles incoming
// motion objects and moves the two servos accordingly.
type Turret struct {
	imgSize   int
	distance  float64
	sleepTime uint64

	x       sysfs.PWMPinner
	y       sysfs.PWMPinner
	adaptor *raspi.Adaptor
}

// New creates a new turret. The pin for each servo is defined by pinX
// and pinY. Distance will be used to make the calculations of the angles
// that would need to be specified. imgSize is the size of the image being
/// processed.
// SleepTime is the amount of time the turret will wait between one movement
// and another one. Note that if this is too low then you might cause some
// damage to the servos.
func New(pinX, pinY string, distance float64, imgSize int, sleepTime uint64) (*Turret, error) {
	r := raspi.NewAdaptor()
	sx, err := r.PWMPin(pinX)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not init pin %s", pinX)
	}
	sy, err := r.PWMPin(pinY)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not init pin %s", pinY)
	}

	if err = r.Connect(); err != nil {
		return nil, errors.Wrap(err, "Could not connect to raspi adaptor")
	}

	t := &Turret{x: sx, y: sy, distance: distance, imgSize: imgSize}
	t.MoveX(0)
	t.MoveY(0)
	return t, nil
}

// calcDutyCycle calculates the duty cycle according
// to the specified angle.
func calcDutyCycle(base uint8) uint32 {
	if base > 180 {
		base = 180
	}
	dc := uint32((float32(base)/180)*(dcMax-dcMin)) + dcMin
	return dc
}

// MoveX moves the servo in the X axis.
func (t *Turret) MoveX(angle uint8) {
	t.x.SetDutyCycle(calcDutyCycle(angle))
}

// MoveY moves the servo in the Y axis.
func (t *Turret) MoveY(angle uint8) {
	t.y.SetDutyCycle(calcDutyCycle(angle))
}

// HandleMotion implements the detector.HandleMotion function.
// This will esentially the heart of the turret. When the detector
// detects motion it will call this function, this will translate
// the detected rectangle into the angles we need in order to move
// both servos to the correct position.
func (t *Turret) HandleMotion(rect image.Rectangle) {
	now := uint64(time.Now().Unix() * 1000)
	if (now - lastRun) < t.sleepTime {
		return
	}
	midX, midY = rectMiddle(rect)
	if lastX == midX && lastY == midY {
		return
	}
	lastRun = now
	lastX, lastY = midX, midY
	x := angleFromPixel(midX, t.imgSize, t.distance) + minX
	y := angleFromPixel(t.imgSize-midY, t.imgSize, t.distance) - minY
	log.Printf("pixels(x,y)=(%v,%v) -- angles(x,y)=(%v,%v)", midX, midY, x, y)
	t.MoveY(y)
	t.MoveX(x)
}

// rectMiddle calculates the middle x and y of a rectangle.
func rectMiddle(rect image.Rectangle) (x int, y int) {
	return (rect.Max.X-rect.Min.X)/2 + rect.Min.X, (rect.Max.Y-rect.Min.Y)/2 + rect.Min.Y
}

func angleFromPixel(pixel, size int, distance float64) uint8 {
	return uint8((math.Atan((float64(pixel)*distance)/float64(size)) * 180 / math.Pi))
}
