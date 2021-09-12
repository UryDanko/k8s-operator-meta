VERSION=0.1
APP=metacontroller
CONTAINER=ydanko/${APP}:${VERSION}
CONTROLLER_VERSION=v1

init-metacontroller:
	kubectl apply -k https://github.com/metacontroller/metacontroller/manifests/production

deploy-metacontroller:
	kubectl apply -k ${CONTROLLER_VERSION}

deploy-sandbox:
	kubectl apply -f sandbox.yaml

undeploy-sandbox:
	kubectl delete -f sandbox.yaml

undeploy-sandbox-controller:
	kubectl delete -f manifests/sandbox-controller.yaml

restart: docker-debug-build docker-push undeploy-sandbox-controller deploy-metacontroller deploy-sandbox

undeploy-metacontroller:
	kubectl delete -k ${CONTROLLER_VERSION}

install:
	go get -v

build:
	GOTRACEBACK=all go build -gcflags "all=-N -l" -o metacontroller

docker-build:
	docker build  -t $(CONTAINER) -f Dockerfile .

docker-debug-build:
	docker build -f Dockerfile.debug -t $(CONTAINER) .

docker-push:
	docker push $(CONTAINER)

