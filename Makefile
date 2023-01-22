# Copyright Konstantin Bakanov 2023

ROOT_DIR    := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

# we are rebuilding regardless of any files changed
all : 
	mkdir -p $(ROOT_DIR)/bin
	go build -o $(ROOT_DIR)/bin/metrics-store $(ROOT_DIR)/cmd/main.go

