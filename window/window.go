package window

import (
	"image"

	"gocv.io/x/gocv"
)

// Manager holds the three windows.
type Manager struct {
	frames     *gocv.Window
	deltas     *gocv.Window
	thresholds *gocv.Window

	img gocv.Mat

	frameSize image.Point
}

// New creates a new window manager where the main
// window will have the specific width and height
func New(width, height int) *Manager {
	f, d, t := gocv.NewWindow("Frames"), gocv.NewWindow("Deltas"), gocv.NewWindow("Thresholds")
	f.ResizeWindow(width, height)
	d.ResizeWindow(400, 400)
	t.ResizeWindow(400, 400)
	f.MoveWindow(0, 0)
	d.MoveWindow(800, 0)
	t.MoveWindow(800, 600)
	return &Manager{
		frames:     f,
		deltas:     d,
		thresholds: t,
		frameSize:  image.Point{X: width, Y: height},
		img:        gocv.NewMat(),
	}
}

// StreamFrame implements the streamer interface.
func (m *Manager) StreamFrame(img gocv.Mat) {
	gocv.Resize(img, &m.img, m.frameSize, 0, 0, gocv.InterpolationLinear)
	imShow(m.frames, m.img)
}

// StreamDelta implements the streamer interface.
func (m *Manager) StreamDelta(img gocv.Mat) {
	imShow(m.deltas, img)
}

// StreamThresh implements the streamer interface.
func (m *Manager) StreamThresh(img gocv.Mat) {
	imShow(m.thresholds, img)
}

func imShow(w *gocv.Window, img gocv.Mat) {
	w.IMShow(img)
	w.WaitKey(1)
}

// Close closes all the windows and mat.
func (m *Manager) Close() error {
	err := m.frames.Close()
	err = m.deltas.Close()
	err = m.thresholds.Close()
	err = m.img.Close()
	return err
}
