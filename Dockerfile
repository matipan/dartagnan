FROM denismakogon/gocv-alpine:3.4.2-buildstage as build-stage

RUN go get -u github.com/golang/dep/cmd/dep
ADD . $GOPATH/src/github.corp.globant.com/InternetOfThings/face-tracking-turret
WORKDIR $GOPATH/src/github.corp.globant.com/InternetOfThings/face-tracking-turret
RUN dep ensure -v
RUN go build -o $GOPATH/bin/face-tracking-turret

FROM denismakogon/gocv-alpine:3.4.2-runtime

COPY --from=build-stage /go/bin/face-tracking-turret /face-tracking-turret

EXPOSE 8080

ENTRYPOINT /face-tracking-turret -device=$DEVICE_ID -area=$MIN_AREA -stream=$STREAM -port=:8080
