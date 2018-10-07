build:
	docker build -t iot/face-tracking-turret .
run:
	docker run -d -p 6677:8080 -e DEVICE_ID=0 -e MIN_AREA=7000 -e STREAM=true --device=/dev/video0 --name face-tracking-turret -t iot/face-tracking-turret
stop:
	docker rm -f face-tracking-turret

