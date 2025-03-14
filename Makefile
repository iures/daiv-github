PLUGIN_NAME=daiv-github

.PHONY: build install clean

install: build
	cp ./out/$(PLUGIN_NAME).so ~/.daiv/plugins/

build: tidy
	go build -o ./out/$(PLUGIN_NAME).so -buildmode=plugin main.go

tidy: clean
	go mod tidy

clean:
	rm -f ./out/$(PLUGIN_NAME).so
	rm -f ~/.daiv/plugins/$(PLUGIN_NAME).so


test:
	go test -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

test-cover-html:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
