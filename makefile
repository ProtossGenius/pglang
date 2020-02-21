unit: build
	cd unit/ && go test -v -count=1

import:
	go get -u github.com/ProtossGenius/SureMoonNet
	cd $(GOPATH)/src/github.com/ProtossGenius/SureMoonNet && make install


test: unit

build:
	go run ./build.go

clean:

all: test 

