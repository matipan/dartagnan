package turret

import (
	"image"
	"log"
	"math"
	"time"

	"gobot.io/x/gobot/platforms/raspi"
	"gobot.io/x/gobot/sysfs"
)

const (
	// PiBlasterPeriod is the PWM period used to configure
	// pi blaster.
	PiBlasterPeriod = 20000000

	maxDutyCycle         = 2350000
	minDutyCycle         = 450000
	minX                 = 70
	minY                 = 10
	distance     float64 = 1.35
)

var (
	lastRun, currentRun int64
	lastX, lastY        int
	midX, midY          int
	x, y                uint8
)

// Turret is the tracking turret
type Turret struct {
	x        sysfs.PWMPinner
	y        sysfs.PWMPinner
	adaptor  *raspi.Adaptor
	size     float64
	distance float64
	sleep    int64
}

// New creates a new turret with connecting the Pi adaptor to
// the servos on the specified pin. For performing the calculations
// that determine the angles for X and Y we use the `distance` and
// imgSize parameters. sleepTime specifies how long to wait between
// one movement and another(in milliseconds). Note that setting this
// too low will definitely damage the servos if they can't move fast
// enough.
func New(pinX, pinY string, distance float64, imgSize, sleepTime int) (*Turret, error) {
	r := raspi.NewAdaptor()
	r.PiBlasterPeriod = PiBlasterPeriod

	sx, err := r.PWMPin(pinX)
	if err != nil {
		return nil, err
	}
	sy, err := r.PWMPin(pinY)
	if err != nil {
		return nil, err
	}

	if err = r.Connect(); err != nil {
		return nil, err
	}
	t := &Turret{x: sx, y: sy, adaptor: r, size: float64(imgSize), sleep: int64(sleepTime)}
	t.MoveX(0)
	t.MoveY(0)
	return t, nil
}

func calcDutyCycle(base uint8) uint32 {
	if base > 180 {
		base = 180
	}
	angle := uint32((float32(base)/180)*(maxDutyCycle-minDutyCycle)) + minDutyCycle
	return angle
}

// MoveX moves the servo of the X axis.
func (t *Turret) MoveX(angle uint8) {
	t.x.SetDutyCycle(calcDutyCycle(angle))
}

// MoveY moves the servo of the Y axis.
func (t *Turret) MoveY(angle uint8) {
	t.y.SetDutyCycle(calcDutyCycle(angle))
}

// HandleMotion handles an incoming rectangle.
func (t *Turret) HandleMotion(rect image.Rectangle) {
	currentRun = time.Now().Unix() * 1000
	if (currentRun - lastRun) < t.sleep {
		return
	}
	midX, midY = middle(rect)
	if lastX == midX && lastY == midY {
		return
	}
	lastX, lastY = midX, midY
	x = angle(midX, minX, t.distance, t.size) + minX
	y = angle(midY-500, minY, t.distance, t.size) - minY
	log.Printf("pixels(x,y)=(%v,%v) -- angles(x,y)=(%v,%v)", int(midX), int(midY), x, y)
	t.MoveY(y)
	t.MoveX(x)
	lastRun = currentRun
}

func middle(r image.Rectangle) (x int, y int) {
	return (r.Max.X-r.Min.X)/2 + r.Min.X, (r.Max.Y-r.Min.Y)/2 + r.Min.Y
}

// angle returns the angle specific for the pixel, size, base and distance
func angle(p int, b uint8, d, s float64) uint8 {
	return uint8((math.Atan((float64(p)*d)/s))*180.0/math.Pi) + b
}
