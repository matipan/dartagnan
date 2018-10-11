package window

import (
	"image"

	"gocv.io/x/gocv"
)

// Manager has three windows, one for each
// type of image.
type Manager struct {
	delta  *gocv.Window
	frame  *gocv.Window
	thresh *gocv.Window
	img    gocv.Mat
	point  image.Point
}

// New creates a new window manager.
func New(sizeX, sizeY int) *Manager {
	frames := gocv.NewWindow("Frame")
	deltas := gocv.NewWindow("Delta")
	threshs := gocv.NewWindow("Threshs")
	return &Manager{
		delta:  deltas,
		thresh: threshs,
		frame:  frames,
		img:    gocv.NewMat(),
		point:  image.Point{X: sizeX, Y: sizeY},
	}
}

// StreamFrame implements the streamer interface.
func (m *Manager) StreamFrame(img gocv.Mat) {
	gocv.Resize(img, &m.img, m.point, 0, 0, gocv.InterpolationLinear)
	imShow(m.frame, m.img)
}

// StreamThresh implements the streamer interface.
func (m *Manager) StreamThresh(img gocv.Mat) {
	imShow(m.thresh, img)
}

// StreamDelta implements the streamer interface.
func (m *Manager) StreamDelta(img gocv.Mat) {
	imShow(m.delta, img)
}

func imShow(w *gocv.Window, img gocv.Mat) {
	w.IMShow(img)
	w.WaitKey(1)
}

// Close closes the windows.
func (m *Manager) Close() error {
	err := m.frame.Close()
	err = m.delta.Close()
	err = m.thresh.Close()
	err = m.img.Close()
	return err
}
