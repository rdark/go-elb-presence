go-elb-presence
==============================

Presence sidekick to add an instance as a backend for an ELB, and optionally a
security group to an instance at the same time.

## Build Instructions

    $ go get .
    $ CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' .
    $ docker build --rm -t $tag_name .

## Sample IAM Policy

Principal of least privilege policy:

    {
      "Version": "2012-10-17",
      "Statement": [{
        "Effect": "Allow",
        "Action": [
          "elasticloadbalancing:DeregisterInstancesFromLoadBalancer",
          "elasticloadbalancing:RegisterInstancesWithLoadBalancer"
        ],
        "Resource": "arn:aws:elasticloadbalancing:eu-wetst-1:ACCOUNTNUMBER:loadbalancer/my-load-balancer"
      }]
    }
