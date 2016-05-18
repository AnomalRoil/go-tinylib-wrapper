EXEC=tinygarble
SOURCE=tinygarble.go

all:        ${SOURCE} 
	    go build  -o ${EXEC} ${SOURCE}

fmt:        ${SOURCE}
	    gofmt -w ${SOURCE}

.PHONY:     install clean

install:
	    go install ./...

clean:
	    rm -rf $(EXEC)
