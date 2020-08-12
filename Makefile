.PHONY: container-linux-userdata-validator
container-linux-userdata-validator:
	go build -mod=vendor .
