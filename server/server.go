package server

import (
	"log"

	"github.com/gorilla/mux"
	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

// Server is a streaming server.
type Server struct {
	*mux.Router
	delta  *mjpeg.Stream
	frame  *mjpeg.Stream
	thresh *mjpeg.Stream
}

// StreamDelta implements the detector.Streamer interface.
func (s *Server) StreamDelta(img gocv.Mat) {
	IMStream(s.delta, img)
}

// StreamThresh implements the detector.Streamer interface.
func (s *Server) StreamThresh(img gocv.Mat) {
	IMStream(s.thresh, img)
}

// StreamFrame implements the detector.Streamer interface.
func (s *Server) StreamFrame(img gocv.Mat) {
	IMStream(s.frame, img)
}

// New creates a new streaming server.
func New() *Server {
	fs := mjpeg.NewStream()
	ds := mjpeg.NewStream()
	ts := mjpeg.NewStream()

	r := mux.NewRouter()
	r.Handle("/frame", fs)
	r.Handle("/delta", ds)
	r.Handle("/thresh", ts)

	return &Server{
		Router: r,
		delta:  ds,
		frame:  fs,
		thresh: ts,
	}
}

// IMStream streams img to the stream.
func IMStream(stream *mjpeg.Stream, img gocv.Mat) {
	buf, err := gocv.IMEncode(".jpg", img)
	if err != nil {
		log.Printf("Could not encode img: %s", err)
		return
	}
	stream.UpdateJPEG(buf)
}
