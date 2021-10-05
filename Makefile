VERSION=0.1
APP=metacontroller
CONTAINER=ydanko/${APP}:${VERSION}
CONTROLLER_VERSION=v1

# ---- Init and deploy -------
init-metacontroller:
	kubectl apply -k https://github.com/metacontroller/metacontroller/manifests/production

deploy-metacontroller:
	kubectl apply -k ${CONTROLLER_VERSION}

deploy-sandbox:
	kubectl apply -f sandbox.yaml

restart: docker-debug-build docker-push deploy-metacontroller deploy-sandbox

# ----- Undeploy ----
undeploy-sandbox:
	kubectl delete -f sandbox.yaml

undeploy-sandbox-controller:
	kubectl delete -f manifests/sandbox-controller.yaml

undeploy-metacontroller:
	kubectl delete -k ${CONTROLLER_VERSION}

install:
	go get -v

# ------ Docker ------
build:
	GOTRACEBACK=all go build -gcflags "all=-N -l" -o metacontroller

docker-build:
	docker build  -t $(CONTAINER) -f Dockerfile .

docker-debug-base-build:
	docker build  -t golang-delve:latest -f Dockerfile.delve .

docker-debug-build:
	docker build -f Dockerfile.debug -t $(CONTAINER) .

docker-push:
	docker push $(CONTAINER)

