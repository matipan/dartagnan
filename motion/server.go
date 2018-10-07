package motion

import (
	"log"

	"github.com/gorilla/mux"
	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

type Server struct {
	*mux.Router
	delta  *mjpeg.Stream
	frame  *mjpeg.Stream
	thresh *mjpeg.Stream
}

func (s *Server) StreamDelta(img gocv.Mat) {
	IMStream(s.delta, img)
}

func (s *Server) StreamThresh(img gocv.Mat) {
	IMStream(s.thresh, img)
}

func (s *Server) StreamFrame(img gocv.Mat) {
	IMStream(s.frame, img)
}

func NewServer() *Server {
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

func IMStream(stream *mjpeg.Stream, img gocv.Mat) {
	buf, err := gocv.IMEncode(".jpg", img)
	if err != nil {
		log.Printf("Could not encode img: %s", err)
		return
	}
	stream.UpdateJPEG(buf)
}
