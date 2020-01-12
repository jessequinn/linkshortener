#!/bin/bash

CGO_ENABLED=1 GOOS=linux go build -o build/app -a -installsuffix cgo -ldflags '-w'
