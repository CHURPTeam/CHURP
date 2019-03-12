#!/usr/bin/env bash

docker build -f ./Dockerfile-builder -t churp/builder .
docker push churp/builder

docker build -f ./Dockerfile-runtime -t churp/runtime .
docker push churp/runtime