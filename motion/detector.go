package motion

import (
	"context"
	"image"
	"image/color"

	"github.com/pkg/errors"
	"gocv.io/x/gocv"
)

var (
	rectColor   = color.RGBA{R: 0, G: 255, B: 0, A: 0}
	textColor   = color.RGBA{R: 0, G: 0, B: 255, A: 0}
	statusPoint = image.Pt(10, 20)
)

type Detector struct {
	video *gocv.VideoCapture

	firstFrame gocv.Mat
	frame      gocv.Mat
	gray       gocv.Mat
	delta      gocv.Mat
	thresh     gocv.Mat
	kernel     gocv.Mat

	cf ContourFunc

	streamer Streamer

	area float64
}

type Streamer interface {
	StreamDelta(img gocv.Mat)
	StreamFrame(img gocv.Mat)
	StreamThresh(img gocv.Mat)
}

type ContourFunc func(rect image.Rectangle, frame gocv.Mat)

func NewDetector(deviceID int, area float64, cf ContourFunc, streamer Streamer) (*Detector, error) {
	video, err := gocv.VideoCaptureDevice(deviceID)
	if err != nil {
		return nil, errors.Wrap(err, "Could not open capture device")
	}

	frame := gocv.NewMat()
	firstFrame := gocv.NewMat()
	if !video.Read(&frame) {
		return nil, errors.Wrap(err, "Could not read first video frame")
	}
	convertFrame(frame, &firstFrame)

	return &Detector{
		video:      video,
		firstFrame: firstFrame,
		frame:      frame,
		gray:       gocv.NewMat(),
		delta:      gocv.NewMat(),
		thresh:     gocv.NewMat(),
		kernel:     gocv.NewMat(),
		streamer:   streamer,
		cf:         cf,
		area:       area,
	}, nil
}

func (d *Detector) Run(ctx context.Context) {
	defer d.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if d.forward() {
				return
			}
		}
	}
}

func (d *Detector) forward() bool {
	if !d.video.Read(&d.frame) {
		return true
	}

	convertFrame(d.frame, &d.gray)

	gocv.AbsDiff(d.firstFrame, d.gray, &d.delta)
	gocv.Threshold(d.delta, &d.thresh, 25, 255, gocv.ThresholdBinary)
	gocv.Dilate(d.thresh, &d.thresh, d.kernel)
	cnt := bestContour(d.thresh.Clone(), d.area)
	if len(cnt) > 0 {
		rect := gocv.BoundingRect(cnt)
		gocv.Rectangle(&d.frame, rect, rectColor, 2)
		gocv.PutText(&d.frame, "Motion detected", statusPoint, gocv.FontHersheyPlain, 1.2, textColor, 2)
		d.cf(rect, d.frame)
	}

	d.streamer.StreamFrame(d.frame)
	d.streamer.StreamDelta(d.delta)
	d.streamer.StreamThresh(d.thresh)

	return false
}

func (d *Detector) Close() error {
	d.firstFrame.Close()
	d.frame.Close()
	d.gray.Close()
	d.delta.Close()
	d.thresh.Close()
	d.kernel.Close()
	return d.video.Close()
}

func bestContour(frame gocv.Mat, minArea float64) []image.Point {
	cnts := gocv.FindContours(frame, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	var (
		bestCnt  []image.Point
		bestArea float64 = minArea
	)
	for _, cnt := range cnts {
		if area := gocv.ContourArea(cnt); area > bestArea {
			bestArea = area
			bestCnt = cnt
		}
	}
	return bestCnt
}

func convertFrame(src gocv.Mat, dst *gocv.Mat) {
	gocv.Resize(src, &src, image.Point{X: 500, Y: 500}, 0, 0, gocv.InterpolationLinear)
	gocv.CvtColor(src, dst, gocv.ColorBGRToGray)
	gocv.GaussianBlur(*dst, dst, image.Point{X: 21, Y: 21}, 0, 0, gocv.BorderReflect101)
}
