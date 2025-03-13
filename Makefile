.PHONY: build install clean

install: build
	cp ./out/daiv-github.so ~/.daiv/plugins/

build: tidy
	go build -o ./out/daiv-github.so -buildmode=plugin main.go

tidy: clean
	go mod tidy

clean:
	rm -f ./out/daiv-github
	rm -f ~/.daiv/plugins/daiv-github.so
