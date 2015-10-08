all: test deploy 

deploy:
	godep go build -v .

test:
	godep go test -v ./...

update-docker-godeps:
	@echo this will overwrite large chunks of your GOPATH. hit enter to continue.
	@read
	go get -d -u github.com/docker/docker || exit 0
	cd "$$(echo $$GOPATH | awk -F: '{ print $$1 }')/src/github.com/docker/docker" && sh hack/make/.go-autogen
	godep restore ./...

clean:
	go clean -i -x ./...
