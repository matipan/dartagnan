build:
	docker build -t iot/face-tracking-turret .
run:
	docker run -d -p 6677:8080 --device=/dev/video0 --name face-tracking-turret -t iot/face-tracking-turret

