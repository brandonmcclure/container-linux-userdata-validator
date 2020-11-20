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

## Contributor Code of Conduct

In the interest of promoting a fair, diverse, and open community, Kinvolk uses
its [Code of Conduct](https://github.com/kinvolk/contribution/blob/master/CODE_OF_CONDUCT.md) for all its projects and events.

Please read and uphold this code-of-conduct while participating in Kinvolk
projects and events.

