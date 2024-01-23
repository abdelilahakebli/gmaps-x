PROJECT_NAME := "gmaps-x"
PROJECT_VERSION := "0.1.0"


cli: buildcli
	@./bin/cli $(ARGS)

api: buildapi
	@./bin/api $(ARGS)

buildcli:
	@go build -o bin/cli cmd/cli/*.go

buildapi:
	@go build -o bin/api cmd/rest/*.go

build:
	@go build -o bin/cli cmd/cli/*.go
	@go build -o bin/api cmd/rest/*.go
	
