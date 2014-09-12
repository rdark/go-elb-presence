package main

import (
	"flag"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/elb"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var lbname, region, accesskey, secretkey string

func init() {
	flag.StringVar(&lbname, "lbname", os.Getenv("ELB_NAME"), "name of the ELB to register the instance with")
	flag.StringVar(&region, "region", os.Getenv("AWS_REGION"), "AWS region in which the ELB resides")
	flag.StringVar(&accesskey, "accesskey", os.Getenv("AWS_ACCESS_KEY"), "AWS Access Key")
	flag.StringVar(&secretkey, "secretkey", os.Getenv("AWS_SECRET_KEY"), "AWS Secret Key")
	flag.Parse()
}

func main() {
	instanceID := aws.InstanceId()
	if instanceID == "unknown" {
		log.Fatalln("Unable to get instance id")
	}

	auth, err := aws.GetAuth(accesskey, secretkey, "", time.Time{})
	if err != nil {
		log.Fatalln("Unable to get AWS auth", err)
	}

	awselb := elb.New(auth, aws.GetRegion(region))
	_, err = awselb.RegisterInstancesWithLoadBalancer([]string{instanceID}, lbname)
	if err != nil {
		log.Fatalln("Error registering instance", err)
	}

	log.Printf("Registered instance %s with elb %s\n", instanceID, lbname)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	// this waits until we get a kill signal
	<-c

	_, err = awselb.DeregisterInstancesFromLoadBalancer([]string{instanceID}, lbname)
	if err != nil {
		log.Fatalln("Error deregistering instance", err)
	}

	log.Printf("Deregistered instance %s with elb %s\n", instanceID, lbname)
}
