
tag := 1.1.7
registries := docker.vpclub.cn
group := hidevopsio
image_name := hinode-java-jar
binary := hinode

dockerimage := $(registries)/$(group)/$(image_name):$(tag)

all : build-linux docker-build docker-push clear

build-linux :
	@echo "build"
	GOOS=linux go build -o $(binary)

gofmt :
	go fmt ./...

govet :
	go vet -v ./...

gotest :
	go test -v

docker-build :
	@echo "docker build -t $(dockerimage) ."
	docker build -t $(dockerimage) .

docker-push :
	@echo "docker build -t $(dockerimage) ."
	docker push $(dockerimage)

clear :
	rm -rf $(binary)
