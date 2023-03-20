#!/bin/bash

#go get -insecure github.com/nurture-farm/Contracts
go mod tidy
go mod vendor

# Build docker image
docker build --no-cache -t core/communication-service:$1 .
