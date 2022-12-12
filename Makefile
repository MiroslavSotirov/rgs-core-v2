# Go parameters
GIT_COMMIT=$(shell git describe --tags)
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run
CONTAINER_NAME=rgs-core
PACKAGE_NAME=gitlab.maverick-ops.com/maverick/rgs-core-v2
IMAGE=harbor.inf01.activeops.io/maverick/mvg_rgs
ifndef BUILDVERSION
	BUILDVERSION=latest
endif

# Dockerfile is not set up to run locally
# add config file copy commands to enable this,
# however this would conflict with the hosted setup

.PHONY: build test start push run stop latest test_all
all: test build docker rundocker

test:
	go test ./... -short

test_all:
	go test ./...

start:
	$(GOBUILD) -v ./...
	$(GORUN) $(PACKAGE_NAME)/cmd -logtostderr=true

runvt:
	$(GOBUILD) -v ./...
	$(GORUN) $(PACKAGE_NAME)/cmd -vt=true -logtostderr=true -engine=mvgEngineXVIII -spins=100000 -chunks=5

push:
	docker push $(IMAGE):$(BUILDVERSION)

latest:
	docker tag "$(IMAGE):$(BUILDVERSION)" "$(IMAGE):latest"
	docker push "$(IMAGE):latest"

docker:
	@echo 'BUILDVERSION set to $(BUILDVERSION)'
	docker build  \
		--pull -t "$(IMAGE):$(BUILDVERSION)" \
		--file Dockerfile .
		
rundocker:
	@echo 'BUILDVERSION set to $(BUILDVERSION)'
	docker run  --name "${CONTAINER_NAME}-$(BUILDVERSION)" -p 3000:3000 -d "$(IMAGE):$(BUILDVERSION)"

stop:
	docker stop "${CONTAINER_NAME}-$(BUILDVERSION)"
	docker rm "${CONTAINER_NAME}-$(BUILDVERSION)"

build:
	$(GOBUILD) -v -ldflags "-X 'main.gitCommit=$(GIT_COMMIT)'" -o rgs cmd/main.go

run:
	./rgs
