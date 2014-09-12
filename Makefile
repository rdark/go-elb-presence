all: go-elb-presence

go-elb-presence: *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a

.PHONY: clean
clean:
	rm go-elb-presence
