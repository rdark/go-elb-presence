FROM busybox

MAINTAINER Bryce Kahle <bryce.kahle@mlssoccer.com>

RUN mkdir -p /etc/ssl/certs
ADD ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ADD ./go-elb-presence /bin/elb-presence

ENTRYPOINT ["/bin/elb-presence"]
