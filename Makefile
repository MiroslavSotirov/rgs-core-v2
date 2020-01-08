# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run
CONTAINER_NAME=rgs-core
PACKAGE_NAME=gitlab.maverick-ops.com/maverick/rgs-core-v2
IMAGE=harbor.maverick-ops.com/maverick/mvg_rgs


.PHONY: build test start push run stop latest test_all
all: test build run

test:
	go test ./... -short

test_all:
	go test ./...

start:
	$(GOBUILD) -v ./...
	$(GORUN) $(PACKAGE_NAME)/cmd -logtostderr=true

runvt:
	$(GOBUILD) -v ./...
	$(GORUN) $(PACKAGE_NAME)/cmd -vt=true -logtostderr=true -engine=RNG

push:
	docker push $(IMAGE):$(BUILDVERSION)

latest:
	docker tag "$(IMAGE):$(BUILDVERSION)" "$(IMAGE):latest"
	docker push "$(IMAGE):latest"

build:
	docker build  \
		--pull -t "$(IMAGE):$(BUILDVERSION)" \
		--file Dockerfile .
run:
	docker run  --name "${CONTAINER_NAME}-$(BUILDVERSION)" -p 3000:3000 -d "$(IMAGE):$(BUILDVERSION)"

stop:
	docker stop "${CONTAINER_NAME}-$(BUILDVERSION)"
	docker rm "${CONTAINER_NAME}-$(BUILDVERSION)"

