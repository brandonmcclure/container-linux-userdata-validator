# Flatcar Container Linux Userdata Validator

![Go](https://github.com/kinvolk/container-linux-userdata-validator/workflows/Go/badge.svg)

This is the code that powered the public service at https://coreos.com/validate/.

## Building

The included multi-stage Dockerfile can be used to build working images. Just run the following:

```shell
docker build .
```

## Updating dependencies

The following commands can be used to update the dependencies of this project:

```shell
go get -u ./...
go mod tidy
```

## Deployment

This repository is configured for autobuilding on Quay, so that new git
tags are automatically available as container images.

Deployable tags are pushed to git in the format `yyyymmdd-rev`.
The corresponding image needs to be manually bumped in the relevant
Helm chart for the CoreOS kubernetes cluster.
