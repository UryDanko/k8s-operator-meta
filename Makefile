VERSION=0.1
APP=metacontroller
CONTAINER=ydanko/${APP}:${VERSION}

install:
	go get -v

docker-build:
	docker build . -t $(CONTAINER)

docker-push:
	docker push $(CONTAINER)

