EXEC=./bin/tinygarble
SOURCE=./example
TLIB=./tinylib

all:        ${SOURCE} 
	    go build  -o ${EXEC} ${SOURCE}
lib:
		go build ${TLIB}
fmt:        ${SOURCE}
	    gofmt -w ${SOURCE}
		gofmt -w ${TLIB}

.PHONY:     install clean test

test:
		go test ./...

install:
	    go install ./... 

clean:
	    rm -rf $(EXEC)
