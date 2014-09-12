all: build/elb-presence

build/elb-presence: *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o build/elb-presence

.PHONY: clean
clean:
	rm build/elb-presence
