# --------------------------------------------------------------------------
# Makefile for the Gas Monetization App
#
# v0.1 (2020/03/09)  - Initial version, base API server build.
# (c) Fantom Foundation, 2023
# --------------------------------------------------------------------------

# project related vars
PROJECT := $(shell basename "$(PWD)")

# go related vars
GO_BASE := $(shell pwd)
GO_BIN := $(CURDIR)/build

# compile time variables will be injected into the app
APP_VERSION := 1.0
BUILD_DATE := $(shell date)
BUILD_COMPILER := $(shell go version)
BUILD_COMMIT := $(shell git show --format="%H" --no-patch)
BUILD_COMMIT_TIME := $(shell git show --format="%cD" --no-patch)

.PHONY: all clean test

all: gas-monetization-app

gas-monetization-app:
	@go build \
    		-ldflags="-X 'ftm-gas-monetization/cmd/gas-monetization-cli/version.Version=$(APP_VERSION)' -X 'ftm-gas-monetization/cmd/gas-monetization-cli/version.Time=$(BUILD_DATE)' -X 'ftm-gas-monetization/cmd/gas-monetization-cli/version.Compiler=$(BUILD_COMPILER)' -X 'ftm-gas-monetization/cmd/gas-monetization-cli/version.Commit=$(BUILD_COMMIT)' -X 'ftm-gas-monetization/cmd/gas-monetization-cli/version.CommitTime=$(BUILD_COMMIT_TIME)'" \
    		-o $(GO_BIN)/gas-monetization \
    		-v \
    		./cmd/gas-monetization-cli

test:
	@go test ./...

clean:
	rm -fr ./build/*
