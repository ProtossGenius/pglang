unit: build
	cd unit/ && go test -v -count=1

import:
	go get -u github.com/ProtossGenius/SureMoonNet
	cd $(GOPATH)/src/github.com/ProtossGenius/SureMoonNet && make install


test: build  unit

build: clean
	go run ./build.go 

clean:
	rm -rf 	./datas/unit/lex_pgl/*.to
	rm -rf ./datas/unit/grm_go/*.to
all: test 
install:
	smdcatalog
rely:
	go get github.com/ProtossGenius/smntools/cmd/smdcatalog
