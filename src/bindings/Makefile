EXEC=bindings
SOURCE=bindings.go

all:
	    go build  -o ${EXEC} ${SOURCE}

fmt:
		gofmt -w ${SOURCE}

.PHONY: install clean

install:
	    go install ./...

clean:
	    rm -rf $(EXEC)
