check:
	go vet ./...
format:
	#go get golang.org/x/tools/cmd/goimports
	find . -name '*.go' | grep -Ev 'vendor|thrift_gen' | xargs goimports -w

build:
	sh ./script/build_api.sh

run:
	sh ./output/run_api.sh

clean:
	rm -rf output

server: format clean build run

docker-build-api:
	docker build -t 172.16.16.172:12380/xxx/xxx:v0.2 -f ./API.Dockerfile ./

docker-push-api:
	docker push 172.16.16.172:12380/xxx/xxx:v0.2

docker-all: clean docker-build-api docker-push-api

# Quick start
# Pull images from dockerhub and run
docker-run-linux:
	sh ./run-for-linux.sh

docker-run-mac:
	sh ./run-for-mac.sh

docker-container-stop:
	docker ps -aq | xargs docker stop
	docker ps -aq | xargs docker rm

docker-image-rm:
	docker image prune --force --all

# Immersive experience
# Compile and run by docker-compose
docker-compose-start:
	docker-compose up -d

docker-compose-stop:
	docker-compose down

docker-compose-build:
	docker-compose build

#USE make TARGET version=xx override version
version ?= latest

docker-tag:
	docker tag schedulx-api:latest galaxyfuture/schedulx-api:${version}

docker-push-hub:
	docker push galaxyfuture/schedulx-api:${version}

docker-hub-all: docker-compose-build docker-tag docker-push-hub