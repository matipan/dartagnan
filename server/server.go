package server

import (
	"errors"
	"image"
	"image/color"
	"net/http"
	"path/filepath"

	"github.com/hybridgroup/mjpeg"
	"github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

type Server struct {
	stream *mjpeg.Stream
	webcam *gocv.VideoCapture
	server *http.Server
}

var (
	logger = logrus.WithField("component", "server")

	ErrInvalidParam     = errors.New("parameter was invalid")
	ErrServerNotStarted = errors.New("called close on a server that has not been started")
)

func (s *Server) Serve(hostport, path string, deviceID int, model, config string) error {
	var err error

	s.webcam, err = gocv.OpenVideoCapture(deviceID)
	if err != nil {
		logger.Errorf("Could not open capture device for deviceID %d: %s", deviceID, err)
		return ErrInvalidParam
	}

	s.stream = mjpeg.NewStream()

	net := gocv.ReadNet(model, config)
	if net.Empty() {
		logger.Errorf("Could not read network model from: %s %s", model, config)
		return ErrInvalidParam
	}

	net.SetPreferableBackend(gocv.NetBackendType(gocv.NetBackendDefault))
	net.SetPreferableTarget(gocv.NetTargetType(gocv.NetTargetCPU))

	var (
		ratio   float64
		mean    gocv.Scalar
		swapRGB bool
	)
	if filepath.Ext(model) == ".caffemodel" {
		ratio = 1.0
		mean = gocv.NewScalar(104, 177, 123, 0)
		swapRGB = false
	} else {
		ratio = 1.0 / 127.5
		mean = gocv.NewScalar(127.5, 127.5, 127.5, 0)
		swapRGB = true
	}

	go s.mjpegCapture(net, ratio, mean, swapRGB)

	s.server = &http.Server{Handler: s.stream, Addr: hostport}
	logger.Infof("Serving on: %s", hostport)
	return s.server.ListenAndServe()
}

func (s *Server) Close() error {
	if s.webcam != nil && s.server != nil {
		s.webcam.Close()
		return s.server.Close()
	}
	return ErrServerNotStarted
}

func (s *Server) mjpegCapture(net gocv.Net, ratio float64, mean gocv.Scalar, swapRGB bool) {
	img := gocv.NewMat()
	defer img.Close()

	for {
		if ok := s.webcam.Read(&img); !ok {
			return
		}
		if img.Empty() {
			continue
		}

		// convert image Mat to 300x300 blob that the object detector can analyze
		blob := gocv.BlobFromImage(img, ratio, image.Pt(300, 300), mean, swapRGB, false)

		// feed the blob into the detector
		net.SetInput(blob, "")

		// run a forward pass thru the network
		prob := net.Forward("")

		performDetection(&img, prob)

		prob.Close()
		blob.Close()

		buf, _ := gocv.IMEncode(".jpg", img)
		s.stream.UpdateJPEG(buf)
	}
}

// performDetection analyzes the results from the detector network,
// which produces an output blob with a shape 1x1xNx7
// where N is the number of detections, and each detection
// is a vector of float values
// [batchId, classId, confidence, left, top, right, bottom]
func performDetection(frame *gocv.Mat, results gocv.Mat) {
	for i := 0; i < results.Total(); i += 7 {
		confidence := results.GetFloatAt(0, i+2)
		if confidence > 0.5 {
			left := int(results.GetFloatAt(0, i+3) * float32(frame.Cols()))
			top := int(results.GetFloatAt(0, i+4) * float32(frame.Rows()))
			right := int(results.GetFloatAt(0, i+5) * float32(frame.Cols()))
			bottom := int(results.GetFloatAt(0, i+6) * float32(frame.Rows()))
			gocv.Rectangle(frame, image.Rect(left, top, right, bottom), color.RGBA{0, 255, 0, 0}, 2)
		}
	}
}
