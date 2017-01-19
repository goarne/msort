#!/usr/bin/make
#
# Makefile for Golang projects.
#
# Features:
# - uses github.com/Masterminds/glide to manage dependencies and uses GO15VENDOREXPERIMENT
#
#

PACKAGE=github.com/goarne/msort
BUILD_FOLDER=$(GOPATH)/bin

CONFIG_FILE=msort.config.json
BINARY=msort
CONFIG_FOLDER=msort.d

CLIENT_CONFIG_FILE=appconfig.yaml
CLIENT_BINARY=msortclient
CLIENT_CONFIG_FOLDER=msortclient.d

WEB_CONFIG_FILE=appconfig.json
WEB_CONFIG_FOLDER=msortweb.d
WEB_BINARY=msortweb

DOCKER_IMAGE=goarne/msortweb
DOCKER_IMAGE_VERSION=latest
DOCKER_CONTAINER=msortweb
#DOCKER_CONTAINER_RUNNING=$(docker ps | grep $(DOCKER_CONTAINER))

test:
	go test ./...	
	
$(BINARY): test
	go build -o ./$(BINARY) $(PACKAGE)

$(CLIENT_BINARY): test
	go build -o ./$(CLIENT_BINARY) $(PACKAGE)/client 
	
$(WEB_BINARY): test
	go build -o ./$(WEB_BINARY) $(PACKAGE)/web

build: $(BINARY) $(CLIENT_BINARY) $(WEB_BINARY)

deploy-$(BINARY):
	go build -o $(BUILD_FOLDER)/$(BINARY) $(PACKAGE)
	test -d "$(BUILD_FOLDER)/$(CONFIG_FOLDER)" ||  mkdir "$(BUILD_FOLDER)/$(CONFIG_FOLDER)"
	cp $(CONFIG_FOLDER)/$(CONFIG_FILE) $(BUILD_FOLDER)/$(CONFIG_FOLDER)

deploy-$(CLIENT_BINARY):
	go build -o $(BUILD_FOLDER)/$(CLIENT_BINARY) $(PACKAGE)/client
	test -d "$(BUILD_FOLDER)/$(CLIENT_CONFIG_FOLDER)" ||  mkdir "$(BUILD_FOLDER)/$(CLIENT_CONFIG_FOLDER)"
	cp client/$(CLIENT_CONFIG_FOLDER)/$(CLIENT_CONFIG_FILE) $(BUILD_FOLDER)/$(CLIENT_CONFIG_FOLDER)
	cp client/$(CLIENT_CONFIG_FOLDER)/$(CONFIG_FILE) $(BUILD_FOLDER)/$(CLIENT_CONFIG_FOLDER)
	
deploy-$(WEB_BINARY):
	go build -o $(BUILD_FOLDER)/$(WEB_BINARY) $(PACKAGE)/web
	test -d "$(BUILD_FOLDER)/$(WEB_CONFIG_FOLDER)" ||  mkdir "$(BUILD_FOLDER)/$(WEB_CONFIG_FOLDER)"
	cp web/$(WEB_CONFIG_FOLDER)/$(WEB_CONFIG_FILE) $(BUILD_FOLDER)/$(WEB_CONFIG_FOLDER)


deploy: deploy-$(BINARY) deploy-$(CLIENT_BINARY) deploy-$(WEB_BINARY)

update-dependencies:
	glide up -s -v -u install

clean: 
	rm -rf $(BUILD_FOLDER)/$(CONFIG_FOLDER) $(BUILD_FOLDER)/$(BINARY) $(BINARY) ./$(CONFIG_FOLDER)/logs
	rm -rf $(BUILD_FOLDER)/$(CLIENT_CONFIG_FOLDER) $(BUILD_FOLDER)/$(CLIENT_BINARY) $(CLIENT_BINARY) client/$(CLIENT_CONFIG_FOLDER)/logs
	rm -rf $(BUILD_FOLDER)/$(WEB_CONFIG_FOLDER) $(BUILD_FOLDER)/$(WEB_BINARY) $(WEB_BINARY) web/$(WEB_CONFIG_FOLDER)/logs
	
	go clean $(PACKAGE)
	go clean $(PACKAGE)/client 
	go clean $(PACKAGE)/web 
	
install:
	glide install
	
docker-build: build
ifneq "$(docker ps | grep $(DOCKER_CONTAINER))" ""
	docker stop $(DOCKER_CONTAINER)
	docker rm $(DOCKER_CONTAINER)
endif

ifneq "$(docker images | grep $(DOCKER_IMAGE):$(DOCKER_IMAGE_VERSION))" ""
	docker rmi $(DOCKER_IMAGE):$(DOCKER_IMAGE_VERSION)
endif
	
	docker build -t $(DOCKER_IMAGE):$(DOCKER_IMAGE_VERSION) .


docker-run: 
ifneq "$(docker ps | grep $(DOCKER_CONTAINER))" ""
	docker stop $(DOCKER_CONTAINER)
	docker rm $(DOCKER_CONTAINER)
endif

	docker run -d -p 8081:8081 --name $(DOCKER_CONTAINER) $(DOCKER_IMAGE):$(DOCKER_IMAGE_VERSION)
