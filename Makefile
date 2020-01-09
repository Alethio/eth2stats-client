VERSION := "$(shell git describe --abbrev=0 --tags 2> /dev/null || echo 'v0.0.0')-$(shell git rev-parse --short HEAD)"

build:
	go build -ldflags "-X main.buildVersion=$(VERSION)"

run:
	go run main.go

install:
	go install -ldflags "-X main.buildVersion=$(VERSION)"

build-docker:
	$(eval KEY:=$(shell cat ~/.ssh/id_rsa_gitlab_machineuser | base64 --wrap 0))
	docker build --build-arg GITLAB_SSH_KEY="${KEY}" -t alethio/eth2stats-client .