EXEC=tinygarble
SOURCE=main.go
TLIB=./tinylib

all:        ${SOURCE} 
	    go build  -o ${EXEC} ${SOURCE}
lib:
		go build ${TLIB}
fmt:        ${SOURCE}
	    gofmt -w ${SOURCE}
test:
		go test ${TLIB}
		go test
.PHONY:     install clean

install:
	    go install ./...

clean:
	    rm -rf $(EXEC)
