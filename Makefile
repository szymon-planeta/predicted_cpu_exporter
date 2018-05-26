build:
	gofmt -e main.go
	docker build . -t predicted_cpu_exporter

tag:
	docker tag predicted_cpu_exporter docker.io/cptplaneta/predicted-cpu-exporter	

push:
	docker push docker.io/cptplaneta/predicted-cpu-exporter

all:	build tag push
