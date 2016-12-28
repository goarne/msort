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


build:
	go test ./...
	go build -o ./$(BINARY) $(PACKAGE)
	go build -o ./$(CLIENT_BINARY) $(PACKAGE)/client 
	go build -o ./$(WEB_BINARY) $(PACKAGE)/web 
	
deploy:	
	go build -o $(BUILD_FOLDER)/$(BINARY) $(PACKAGE)
	test -d "$(BUILD_FOLDER)/$(CONFIG_FOLDER)" ||  mkdir "$(BUILD_FOLDER)/$(CONFIG_FOLDER)"
	cp $(CONFIG_FOLDER)/$(CONFIG_FILE) $(BUILD_FOLDER)/$(CONFIG_FOLDER)

	go build -o $(BUILD_FOLDER)/$(CLIENT_BINARY) $(PACKAGE)/client
	test -d "$(BUILD_FOLDER)/$(CLIENT_CONFIG_FOLDER)" ||  mkdir "$(BUILD_FOLDER)/$(CLIENT_CONFIG_FOLDER)"
	cp client/$(CLIENT_CONFIG_FOLDER)/$(CLIENT_CONFIG_FILE) $(BUILD_FOLDER)/$(CLIENT_CONFIG_FOLDER)
	cp client/$(CLIENT_CONFIG_FOLDER)/$(CONFIG_FILE) $(BUILD_FOLDER)/$(CLIENT_CONFIG_FOLDER)
	
	go build -o $(BUILD_FOLDER)/$(WEB_BINARY) $(PACKAGE)/web
	test -d "$(BUILD_FOLDER)/$(WEB_CONFIG_FOLDER)" ||  mkdir "$(BUILD_FOLDER)/$(WEB_CONFIG_FOLDER)"
	cp web/$(WEB_CONFIG_FOLDER)/$(WEB_CONFIG_FILE) $(BUILD_FOLDER)/$(WEB_CONFIG_FOLDER)

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
		
