FROM denismakogon/gocv-alpine:3.4.2-buildstage as build-stage

RUN go get -u github.com/golang/dep/cmd/dep
ADD . $GOPATH/src/github.corp.globant.com/InternetOfThings/face-tracking-turret
WORKDIR $GOPATH/src/github.corp.globant.com/InternetOfThings/face-tracking-turret
RUN dep ensure -v
RUN go build -o $GOPATH/bin/face-tracking-turret

FROM denismakogon/gocv-alpine:3.4.2-runtime

COPY --from=build-stage /go/bin/face-tracking-turret /face-tracking-turret
ADD ./deploy.prototxt /deploy.prototxt
ADD ./res10_300x300_ssd_iter_140000.caffemodel /res10_300x300_ssd_iter_140000.caffemodel

ENTRYPOINT ["/face-tracking-turret"]
