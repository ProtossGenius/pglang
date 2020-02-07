unit: FORCE
	cd unit/ && go test -v -count=1

import:
	go get -u github.com/ProtossGenius/SureMoonNet

FORCE:

all: unit

